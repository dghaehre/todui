package main

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
)

type DB struct {
	conn *sql.DB
}

func (db DB) Close() error {
	return db.conn.Close()
}

func NewDB(path string) (DB, error) {
	err := creatFileIfNotExist(path)
	if err != nil {
		return DB{}, err
	}
	conn, err := sql.Open("sqlite3", path)
	if err != nil {
		return DB{}, err
	}
	err = conn.Ping()
	if err != nil {
		return DB{}, err
	}
	return DB{
		conn: conn,
	}, nil
}
