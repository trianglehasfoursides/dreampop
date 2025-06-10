package cmd

import (
	"encoding/binary"
	"errors"
	"fmt"
	"log/slog"
	"strconv"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
	"github.com/trianglehasfoursides/dreampop/db"
	"go.etcd.io/bbolt"
)

func init() {
	root.AddCommand(todoAdd)
	root.AddCommand(todoList)
	root.AddCommand(todoEdit)
	root.AddCommand(todoDel)
	root.AddCommand(todoCheck)
	root.AddCommand(todoHistory)
}

func init() {
	// Start a writable transaction.
	tx, err := db.Db.Begin(true)
	if err != nil {
		slog.Error(err.Error())
		return
	}
	defer tx.Rollback()

	// Use the transaction...
	_, err = tx.CreateBucket([]byte("todo"))
	if err != nil {
		return
	}

	// Commit the transaction and check for error.
	if err := tx.Commit(); err != nil {
		slog.Error(err.Error())
		return
	}
}

func init() {
	// Start a writable transaction.
	tx, err := db.Db.Begin(true)
	if err != nil {
		slog.Error(err.Error())
		return
	}
	defer tx.Rollback()

	// Use the transaction...
	_, err = tx.CreateBucket([]byte("todo_history"))
	if err != nil {
		return
	}

	// Commit the transaction and check for error.
	if err := tx.Commit(); err != nil {
		slog.Error(err.Error())
		return
	}
}

// itob returns an 8-byte big endian representation of v.
func itob(v uint64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(v))
	return b
}

var todoAdd = &cobra.Command{
	Use:   "add",
	Short: "",
	Run: func(cmd *cobra.Command, args []string) {
		var todoValue string
		if len(args) > 1 {
			todoValue = args[0]
			db.Db.Update(func(tx *bbolt.Tx) error {
				bucket := tx.Bucket([]byte("todo"))
				id, _ := bucket.NextSequence()
				return bucket.Put(itob(id), []byte(todoValue))
			})
			return
		}
		input := huh.NewInput().
			CharLimit(100).
			Placeholder("hmmm").
			Title("Add new todo").
			Value(&todoValue)
		err := input.Run()
		if err != nil {
			slog.Error(err.Error())
			return
		}

		db.Db.Update(func(tx *bbolt.Tx) error {
			bucket := tx.Bucket([]byte("todo"))
			id, _ := bucket.NextSequence()
			return bucket.Put(itob(id), []byte(todoValue))
		})

	},
}

var todoList = &cobra.Command{
	Use:   "ls",
	Short: "",
	Run: func(cmd *cobra.Command, args []string) {
		db.Db.View(func(tx *bbolt.Tx) error {
			// Assume bucket exists and has keys
			bucket := tx.Bucket([]byte("todo"))
			cursor := bucket.Cursor()

			for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
				key := binary.BigEndian.Uint64(k)
				fmt.Printf("%d. %s \n", key, v)
			}
			return nil
		})
	},
}

var todoEdit = &cobra.Command{
	Use:   "edit",
	Short: "",
	Run: func(cmd *cobra.Command, args []string) {
		var val []byte
		if err := db.Db.View(func(tx *bbolt.Tx) error {
			bucket := tx.Bucket([]byte("todo"))
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

		var todoValue string
		input := huh.NewInput().
			CharLimit(100).
			Placeholder(string(val)).
			Title("Edit todo").
			Value(&todoValue)
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
			return tx.Bucket([]byte("todo")).Put(itob(uint64(key)), []byte(todoValue))
		}); err != nil {
			slog.Error(err.Error())
			return
		}
	},
}

var todoCheck = &cobra.Command{
	Use:   "check",
	Short: "",
	Run: func(cmd *cobra.Command, args []string) {
		var todoKey []byte
		var todoValue []byte
		if err := db.Db.Update(func(tx *bbolt.Tx) error {
			key, err := strconv.Atoi(args[0])
			if err != nil {
				return err
			}
			bucket := tx.Bucket([]byte("todo"))
			val := bucket.Get(itob(uint64(key)))
			if val == nil {
				return errors.New("key doenst exist")
			}
			todoValue = val
			todoKey = itob(uint64(key))
			bucket.Delete(itob(uint64(key)))
			return nil
		}); err != nil {
			slog.Error(err.Error())
			return
		}

		_ = db.Db.Update(func(tx *bbolt.Tx) error {
			bucket := tx.Bucket([]byte("todo_history"))
			bucket.Put(todoKey, todoValue)
			return nil
		})
	},
}

var todoHistory = &cobra.Command{
	Use:   "history",
	Short: "",
	Run: func(cmd *cobra.Command, args []string) {
		db.Db.View(func(tx *bbolt.Tx) error {
			// Assume bucket exists and has keys
			bucket := tx.Bucket([]byte("todo_history"))
			cursor := bucket.Cursor()

			for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
				key := binary.BigEndian.Uint64(k)
				fmt.Printf("%d. %s \n", key, v)
			}
			return nil
		})
	},
}

var todoDel = &cobra.Command{
	Use:   "rm",
	Short: "",
	Run: func(cmd *cobra.Command, args []string) {
		if err := db.Db.Update(func(tx *bbolt.Tx) error {
			key, err := strconv.Atoi(args[0])
			if err != nil {
				return err
			}
			return tx.Bucket([]byte("todo")).Delete(itob(uint64(key)))
		}); err != nil {
			slog.Error(err.Error())
			return
		}
	},
}
