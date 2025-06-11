package cmd

import (
	"encoding/binary"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
	"github.com/trianglehasfoursides/dreampop/db"
	"go.etcd.io/bbolt"
)

func init() {
	root.AddCommand(noteAdd)
	root.AddCommand(noteList)
	root.AddCommand(noteEdit)
	root.AddCommand(noteDel)
	root.AddCommand(noteCheck)
	root.AddCommand(history)
	history.AddCommand(historyClean)
}

// itob returns an 8-byte big endian representation of v.
func itob(v uint64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(v))
	return b
}

var noteAdd = &cobra.Command{
	Use:   "add",
	Short: "",
	Run: func(cmd *cobra.Command, args []string) {
		var noteValue string
		if len(args) > 0 {
			noteValue = args[0]
			if err := db.Db.Update(func(tx *bbolt.Tx) error {
				bucket := tx.Bucket([]byte(self()))
				id, _ := bucket.NextSequence()
				return bucket.Put(itob(id), []byte(noteValue))
			}); err != nil {
				slog.Error(err.Error())
				return
			}
			return
		}
		input := huh.NewInput().
			CharLimit(100).
			Placeholder("hmmm").
			Title("Add new note").
			Value(&noteValue).Validate(func(s string) error {
			if s == "" {
				return errors.New("note can't be empty")
			}
			return nil
		})
		err := input.Run()
		if err != nil {
			slog.Error(err.Error())
			return
		}

		if err = db.Db.Update(func(tx *bbolt.Tx) error {
			bucket := tx.Bucket([]byte(self()))
			id, _ := bucket.NextSequence()
			return bucket.Put(itob(id), []byte(noteValue))
		}); err != nil {
			slog.Error(err.Error())
			return
		}

	},
}

var noteList = &cobra.Command{
	Use:   "ls",
	Short: "",
	Run: func(cmd *cobra.Command, args []string) {
		db.Db.View(func(tx *bbolt.Tx) error {
			// Assume bucket exists and has keys
			bucket := tx.Bucket([]byte(self()))
			cursor := bucket.Cursor()

			for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
				key := binary.BigEndian.Uint64(k)
				fmt.Printf("%d. %s \n", key, v)
			}
			return nil
		})
	},
}

var noteEdit = &cobra.Command{
	Use:   "edit",
	Short: "",
	Run: func(cmd *cobra.Command, args []string) {
		var val []byte
		if err := db.Db.View(func(tx *bbolt.Tx) error {
			bucket := tx.Bucket([]byte(self()))
			key, err := strconv.Atoi(args[0])
			if err != nil {
				return err
			}
			val = bucket.Get(itob(uint64(key)))
			if val == nil {
				return errors.New("key doenst exist")
			}
			return nil
		}); err != nil {
			slog.Error(err.Error())
			return
		}

		var noteValue string
		input := huh.NewInput().
			CharLimit(100).
			Placeholder(string(val)).
			Title("Edit note").
			Value(&noteValue).Validate(func(s string) error {
			if s == "" {
				return errors.New("note can't be empty")
			}
			return nil
		})

		err := input.Run()
		if err != nil {
			slog.Error(err.Error())
			return
		}

		if err = db.Db.Update(func(tx *bbolt.Tx) error {
			key, err := strconv.Atoi(args[0])
			if err != nil {
				return err
			}
			return tx.Bucket([]byte("notes")).Put(itob(uint64(key)), []byte(noteValue))
		}); err != nil {
			slog.Error(err.Error())
			return
		}
	},
}

var noteCheck = &cobra.Command{
	Use:   "check",
	Short: "",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) > 0 {
			if err := db.Db.Update(func(tx *bbolt.Tx) error {
				var vals [][]byte

				for i, _ := range args {
					bucket := tx.Bucket([]byte(self()))
					key, err := strconv.Atoi(args[i])
					if err != nil {
						return err
					}

					val := bucket.Get(itob(uint64(key)))
					if val == nil {
						continue
					}

					vals = append(vals, val)
					bucket.Delete(itob(uint64(key)))
				}

				_ = db.Db.Update(func(tx *bbolt.Tx) error {
					bucket := tx.Bucket([]byte("history"))
					for _, historyVal := range vals {
						id, _ := bucket.NextSequence()
						bucket.Put(itob(id), historyVal)
					}

					return nil
				})
				return nil
			}); err != nil {
				slog.Error(err.Error())
				return
			}
		}

		var items []huh.Option[string]
		var checks []string

		db.Db.View(func(tx *bbolt.Tx) error {
			// Assume bucket exists and has keys
			bucket := tx.Bucket([]byte(self()))
			cursor := bucket.Cursor()

			for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
				key := binary.BigEndian.Uint64(k)
				strKey := strconv.Itoa(int(key))
				items = append(items, huh.NewOption(string(v), strKey+","+string(v)))
			}
			return nil
		})

		multisSelect := huh.NewMultiSelect[string]().
			Title("Check your notes").
			Options(
				items...,
			).
			Value(&checks)

		multisSelect.Run()

		_ = db.Db.Update(func(tx *bbolt.Tx) error {
			bucket := tx.Bucket([]byte(self()))
			for i := range len(checks) {
				valkey := strings.Split(checks[i], ",")
				idkey, _ := strconv.Atoi(valkey[0])
				bucket.Delete(itob(uint64(idkey)))
			}
			return nil
		})

		_ = db.Db.Update(func(tx *bbolt.Tx) error {
			bucket := tx.Bucket([]byte("history"))
			for i := range len(checks) {
				valkey := strings.Split(checks[i], ",")
				id, _ := bucket.NextSequence()
				bucket.Put(itob(id), []byte(valkey[1]))
			}
			return nil
		})
	},
}

var noteDel = &cobra.Command{
	Use:   "rm",
	Short: "",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) > 0 {
			if err := db.Db.Update(func(tx *bbolt.Tx) error {
				var vals [][]byte

				for i, _ := range args {
					bucket := tx.Bucket([]byte(self()))
					key, err := strconv.Atoi(args[i])
					if err != nil {
						return err
					}

					val := bucket.Get(itob(uint64(key)))
					if val == nil {
						continue
					}

					vals = append(vals, val)
					bucket.Delete(itob(uint64(key)))
				}
				return nil
			}); err != nil {
				slog.Error(err.Error())
				return
			}
		}

		var items []huh.Option[string]
		var checks []string

		db.Db.View(func(tx *bbolt.Tx) error {
			// Assume bucket exists and has keys
			bucket := tx.Bucket([]byte(self()))
			cursor := bucket.Cursor()

			for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
				items = append(items, huh.NewOption(string(k), string(v)))
			}
			return nil
		})

		huh.NewMultiSelect[string]().
			Title("Check your notes").
			Options(
				items...,
			).
			Value(&checks)

		_ = db.Db.Update(func(tx *bbolt.Tx) error {
			bucket := tx.Bucket([]byte(self()))
			for _ = range len(checks) {
				bucket.Delete([]byte(checks[0]))
			}
			return nil
		})
	},
}

var history = &cobra.Command{
	Use:   "history",
	Short: "",
	Run: func(cmd *cobra.Command, args []string) {
		db.Db.View(func(tx *bbolt.Tx) error {
			// Assume bucket exists and has keys
			bucket := tx.Bucket([]byte("history"))
			cursor := bucket.Cursor()

			for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
				key := binary.BigEndian.Uint64(k)
				fmt.Printf("%d. %s \n", key, v)
			}
			return nil
		})
	},
}

var historyClean = &cobra.Command{
	Use:   "clean",
	Short: "",
	Run: func(cmd *cobra.Command, args []string) {
		tx, err := db.Db.Begin(true)
		if err != nil {
			slog.Error(err.Error())
			return
		}
		defer tx.Rollback()

		// Use the transaction...
		tx.DeleteBucket([]byte("history"))
		tx.CreateBucket([]byte("history"))

		// Commit the transaction and check for error.
		if err := tx.Commit(); err != nil {
			slog.Error(err.Error())
			os.Exit(1)
		}
	},
}

func self() (self []byte) {
	db.Db.View(func(tx *bbolt.Tx) error {
		self = tx.Bucket([]byte("internal")).Get([]byte("self"))
		return nil
	})
	return
}
