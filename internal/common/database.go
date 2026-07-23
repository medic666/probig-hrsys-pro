package common

import (
	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

var DB *sqlx.DB

func InitDB(dbPath string, wal bool) error {
	dsn := dbPath
	pragmas := "?_journal_mode=WAL&_synchronous=NORMAL&_foreign_keys=ON"
	if wal {
		dsn += pragmas
	}

	var err error
	DB, err = sqlx.Open("sqlite", dsn)
	if err != nil {
		return err
	}

	DB.SetMaxOpenConns(1)
	return DB.Ping()
}

func CloseDB() {
	if DB != nil {
		DB.Close()
	}
}
