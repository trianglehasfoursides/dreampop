package db

import (
	"log/slog"
	"os"

	"github.com/adrg/xdg"
	"go.etcd.io/bbolt"
)

var (
	Db      *bbolt.DB
	err     error
	datadir = xdg.DataHome + "/dreampop"
)

func init() {
	if err := os.MkdirAll(datadir, os.ModePerm); err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
	Db, err = bbolt.Open(datadir+"/dreampop.db", 0600, bbolt.DefaultOptions)
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
}
