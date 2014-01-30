package main

import (
	"bytes"
	"github.com/cznic/kv"
	"path/filepath"
)

var (
	defaultOptions = kv.Options{
		Compare: keyComparator,
	}
)

type DB struct {
	folder string
	metadb *kv.DB
}

func CreateDB(folder string) (*DB, error) {
	db := &DB{folder: folder}
	err := db.createMetaDB()
	return db, err
}

func (db *DB) Close() error {
	if db.metadb != nil {
		return db.metadb.Close()
	}
	return nil
}

func (db *DB) createMetaDB() (error) {
	var err error
	db.metadb, err = kv.Create(filepath.Join(db.folder, "meta.kv"), &defaultOptions)
	return err
}

func keyComparator(a,b []byte) int {
	return bytes.Compare(a,b)
}
