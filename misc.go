package main

import (
	"database/sql"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dsn)
	check(err)

	if err = db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}
