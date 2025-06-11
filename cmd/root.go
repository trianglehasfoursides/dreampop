package cmd

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/spf13/cobra"
	"github.com/trianglehasfoursides/dreampop/db"
)

func init() {
	// Start a writable transaction.
	tx, err := db.Db.Begin(true)
	if err != nil {
		slog.Error(err.Error())
		return
	}
	defer tx.Rollback()

	// Use the transaction...
	_, err = tx.CreateBucketIfNotExists([]byte("notes"))
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	// Commit the transaction and check for error.
	if err := tx.Commit(); err != nil {
		slog.Error(err.Error())
		os.Exit(1)
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
	_, err = tx.CreateBucketIfNotExists([]byte("history"))
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	// Commit the transaction and check for error.
	if err := tx.Commit(); err != nil {
		slog.Error(err.Error())
		os.Exit(1)
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
	_, err = tx.CreateBucketIfNotExists([]byte("internal"))
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	bucket := tx.Bucket([]byte("internal"))
	if self := bucket.Get([]byte("self")); self == nil {
		bucket.Put([]byte("self"), []byte("notes"))
	}

	// Commit the transaction and check for error.
	if err := tx.Commit(); err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
}

var root = &cobra.Command{
	Use:   "dreampop",
	Short: "Your notelist but in terminal",
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
	},
}

func Execute() error {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Error : ", r)
		}
	}()
	if err := root.Execute(); err != nil {
		return err
	}
	return nil
}
