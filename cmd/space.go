package cmd

import (
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
	"github.com/trianglehasfoursides/dreampop/db"
	"go.etcd.io/bbolt"
)

func init() {
	space.AddCommand(spaceAdd)
	space.AddCommand(spaceEdit)
	space.AddCommand(spaceList)
	space.AddCommand(spaceSelect)
	space.AddCommand(spaceDel)
	space.AddCommand(spaceSelf)
	root.AddCommand(space)
}

var space = &cobra.Command{
	Use:   "space",
	Short: "",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

var spaceAdd = &cobra.Command{
	Use:   "add",
	Short: "",
	Run: func(cmd *cobra.Command, args []string) {
		var name string
		if len(args) > 1 {
			name = args[0]
			if name == "internal" || name == "history" {
				err := fmt.Sprintf("can't add space with name '%s',its by design", name)
				slog.Error(err)
				return
			}
			tx, err := db.Db.Begin(true)
			if err != nil {
				slog.Error(err.Error())
				return
			}
			defer tx.Rollback()

			// Use the transaction...
			_, err = tx.CreateBucketIfNotExists([]byte(name))
			if err != nil {
				slog.Error(err.Error())
				return
			}

			// Commit the transaction and check for error.
			if err := tx.Commit(); err != nil {
				slog.Error(err.Error())
				return
			}
		}
		input := huh.NewInput().
			CharLimit(100).
			Placeholder("hmmm").
			Title("Add new space").
			Value(&name).Validate(func(s string) error {
			if s == "" {
				return errors.New("Space can't be empty")
			}
			if s == "internal" || s == "history" {
				return errors.New("can't add,it's by design")
			}
			return nil
		})

		err := input.Run()
		if err != nil {
			slog.Error(err.Error())
			return
		}

		tx, err := db.Db.Begin(true)
		if err != nil {
			slog.Error(err.Error())
			return
		}
		defer tx.Rollback()

		// Use the transaction...
		_, err = tx.CreateBucket([]byte(name))
		if err != nil {
			slog.Error("space already exist")
			return
		}

		// Commit the transaction and check for error.
		if err := tx.Commit(); err != nil {
			slog.Error(err.Error())
			return
		}
	},
}

var spaceEdit = &cobra.Command{
	Use:   "edit",
	Short: "",
	Run: func(cmd *cobra.Command, args []string) {
		var src string
		var dst string
		if len(args) > 1 {
			src = args[0]
			dst = args[1]

			if dst == "internal" || dst == "history" {
				err := fmt.Sprintf("can't edit %s,its by design", dst)
				slog.Error(err)
				return
			}
			tx, err := db.Db.Begin(true)
			if err != nil {
				slog.Error(err.Error())
				return
			}
			defer tx.Rollback()

			srcBucket := tx.Bucket([]byte(src))
			if srcBucket == nil {
				err := errors.New("Source space does not exist")
				slog.Error(err.Error())
				os.Exit(1)
			}

			_, err = tx.CreateBucketIfNotExists([]byte(dst))
			if err != nil {
				slog.Error(err.Error())
				os.Exit(1)
			}

			if err = tx.MoveBucket([]byte(dst), tx.Bucket([]byte(src)), tx.Bucket([]byte(dst))); err != nil {
				slog.Error(err.Error())
				os.Exit(1)
			}

			// Commit the transaction and check for error.
			if err := tx.Commit(); err != nil {
				slog.Error(err.Error())
				return
			}
		}

		form := huh.NewForm(huh.NewGroup(
			huh.NewInput().
				CharLimit(100).
				Placeholder("hmmm").
				Title("Old space").
				Value(&src).Validate(func(s string) error {
				if s == "" {
					return errors.New("Old space can't be empty")
				}
				if s == "internal" || s == "history" {
					return errors.New("can't replace,it's by design")
				}
				return nil
			}),
			huh.NewInput().
				CharLimit(100).
				Placeholder("hmmm").
				Title("New space").
				Value(&src).Validate(func(s string) error {
				if s == "" {
					return errors.New("Old space can't be empty")
				}
				if s == "internal" || s == "history" {
					return errors.New("can't replace,it's by design")
				}
				return nil
			}),
		))

		if err := form.Run(); err != nil {
			slog.Error(err.Error())
			return
		}

		tx, err := db.Db.Begin(true)
		if err != nil {
			slog.Error(err.Error())
			return
		}
		defer tx.Rollback()

		// Use the transaction...
		_, err = tx.CreateBucketIfNotExists([]byte(dst))
		if err != nil {
			slog.Error(err.Error())
			os.Exit(1)
		}

		if err = tx.MoveBucket([]byte(dst), tx.Bucket([]byte(src)), tx.Bucket([]byte(dst))); err != nil {
			slog.Error(err.Error())
			os.Exit(1)
		}

		// Commit the transaction and check for error.
		if err := tx.Commit(); err != nil {
			slog.Error(err.Error())
			return
		}
	},
}

var spaceList = &cobra.Command{
	Use:   "ls",
	Short: "",
	Run: func(cmd *cobra.Command, args []string) {
		tx, err := db.Db.Begin(true)
		if err != nil {
			slog.Error(err.Error())
			return
		}
		defer tx.Rollback()
		tx.ForEach(func(name []byte, b *bbolt.Bucket) error {
			if string(name) == "internal" || string(name) == "history" {
				return nil
			}
			fmt.Printf("%s \n", name)
			return nil
		})
		tx.Commit()
	},
}

var spaceSelect = &cobra.Command{
	Use:   "select",
	Short: "",
	Run: func(cmd *cobra.Command, args []string) {
		var name string
		if len(args) > 0 {
			name = args[0]
			if name == "internal" || name == "history" {
				slog.Error("can't select")
				return
			}
			tx, err := db.Db.Begin(true)
			if err != nil {
				slog.Error(err.Error())
				return
			}
			defer tx.Rollback()

			bucketName := tx.Bucket([]byte(name))
			if bucketName == nil {
				err = errors.New("space does not exist")
				slog.Error(err.Error())
				return
			}

			if err := db.Db.Update(func(tx *bbolt.Tx) error {
				bucket := tx.Bucket([]byte("internal"))
				return bucket.Put([]byte("self"), []byte(name))
			}); err != nil {
				slog.Error(err.Error())
				return
			}
			return
		}

		var spaces []huh.Option[string]

		tx, err := db.Db.Begin(true)
		if err != nil {
			slog.Error(err.Error())
			return
		}
		defer tx.Rollback()
		tx.ForEach(func(name []byte, b *bbolt.Bucket) error {
			if string(name) == "internal" || string(name) == "history" {
				return nil
			}
			spaces = append(spaces, huh.NewOption(string(name), string(name)))
			return nil
		})

		selct := huh.NewSelect[string]().
			Title("Choose your space").
			Options(
				spaces...,
			).
			Value(&name)

		_ = selct.Run()
		fmt.Println(name)

		if err = tx.Bucket([]byte("internal")).Put([]byte("self"), []byte(name)); err != nil {
			slog.Error(err.Error())
			return
		}
		tx.Commit()
	},
}

var spaceDel = &cobra.Command{
	Use:   "rm",
	Short: "",
	Run: func(cmd *cobra.Command, args []string) {
		var name string
		if len(args) > 1 {
			name = args[0]
			if name == "internal" || name == string(self()) || name == "history" {
				slog.Error("can't select")
				return
			}

			tx, err := db.Db.Begin(true)
			if err != nil {
				slog.Error(err.Error())
				return
			}
			defer tx.Rollback()

			bucketName := tx.Bucket([]byte(name))
			if bucketName == nil {
				err = errors.New("space does not exist")
				slog.Error(err.Error())
				return
			}
			tx.DeleteBucket([]byte(name))
			return
		}

		var spaces []huh.Option[string]

		tx, err := db.Db.Begin(true)
		if err != nil {
			slog.Error(err.Error())
			return
		}
		defer tx.Rollback()
		tx.ForEach(func(name []byte, b *bbolt.Bucket) error {
			if string(name) == "internal" {
				return nil
			}
			spaces = append(spaces, huh.NewOption(string(name), string(name)))
			return nil
		})

		selct := huh.NewSelect[string]().
			Title("Choose your space").
			Options(
				spaces...,
			).
			Value(&name)

		selct.Run()

		if name == "internal" || name == string(self()) || name == "history" {
			slog.Error("can't select")
			return
		}

		tx.DeleteBucket([]byte(name))
		tx.Commit()
	},
}

var spaceSelf = &cobra.Command{
	Use:   "self",
	Short: "",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(string(self()))
	},
}
