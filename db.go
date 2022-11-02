package main

import (
	"database/sql"
	"errors"
	_ "github.com/mattn/go-sqlite3"
	"io/ioutil"
	"os"
)

type DB struct {
	conn *sql.DB
}

func (db DB) Close() error {
	return db.conn.Close()
}

func creatFileIfNotExist(path string) error {
	_, err := os.Stat(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return ioutil.WriteFile(path, []byte(""), 0644)
		}
	}
	return err
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
