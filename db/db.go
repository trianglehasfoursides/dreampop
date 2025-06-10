package db

import (
	"log/slog"
	"os"

	"go.etcd.io/bbolt"
)

var (
	Db  *bbolt.DB
	err error
)

func init() {
	Db, err = bbolt.Open("/tmp/dreampop.db", 0600, bbolt.DefaultOptions)
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
}
