package main

import (
	"database/sql"
	"fmt"
	"github.com/cznic/bufs"
	_ "github.com/mattn/go-sqlite3"
	"path/filepath"
)

var (
	minKeywordCode = uint32(1)
	sharedBufs     = &bufs.CCache{}
)

const (
	sqlKeywordTable = ` create table if not exists keywords (
		code integer primary key autoincrement,
		name text not null unique
	)`
	sqlInsertKeyword = `insert into keywords (name) values (?)`
	sqlKeywordByName = `select code, name from keywords where name = ?`
	sqlKeywordByCode = `select code, name form keywords where code = ?`
)

type DB struct {
	folder string
	metadb *sql.DB
}

func CreateDB(folder string) (*DB, error) {
	db := &DB{folder: folder}
	err := db.createMetaDB(filepath.Join(folder, "meta.db"))
	return db, err
}

func (db *DB) createMetaDB(path string) error {
	var err error
	db.metadb, err = sql.Open("sqlite3", path)
	if err != nil {
		return err
	}

	_, err = db.metadb.Exec(sqlKeywordTable)
	if err != nil {
		return err
	}
	return err
}

func (db *DB) Close() error {
	if db.metadb != nil {
		return db.metadb.Close()
	}
	return nil
}

func (db *DB) CreateKeyword(key *Keyword) error {
	if !key.ValidName() {
		return fmt.Errorf("%v is invalid.", key)
	}
	if exists, err := db.keywordExists(key); exists {
		return err
	} else {
		return db.insertKeyword(key)
	}
	panic("not reached")
	return nil
}

func (db *DB) keywordExists(key *Keyword) (bool, error) {
	row := db.metadb.QueryRow(sqlKeywordByName, key.name)
	var code uint64
	var name string
	err := row.Scan(&code, &name)
	key.val = uint32(code)
	return err == nil && key.Valid(), err
}

func (db *DB) insertKeyword(key *Keyword) error {
	result, err := db.metadb.Exec(sqlInsertKeyword, key.name)
	if err != nil {
		return err
	}
	val, err := result.LastInsertId()
	key.val = uint32(val)
	return err
}
