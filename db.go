package main

import (
  "database/sql"

  _ "github.com/go-sql-driver/mysql"
)

const (
  DRIVER   = "mysql"
)

type DBHandler interface {
  Exec(...interface{}) error
  Close()
}

type dbhandler struct {
  db   *sql.DB
  stmt *sql.Stmt
}

func(self *dbhandler) Exec(args ...interface{}) error {
  _, err := self.stmt.Exec(args...)
  return err
}

func(self *dbhandler) Close() {
  self.stmt.Close()
  self.db.Close()
}

func newDb(query string, connStr string) DBHandler {
  db, err := sql.Open(DRIVER, connStr)
	if err != nil {
		panic(err)
	}

  stmt, err := db.Prepare(query)
  if err != nil {
    panic(err)
  }

  return &dbhandler{
    db: db,
    stmt: stmt,
  }
}
