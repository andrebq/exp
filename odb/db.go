package odb

import (
	"bytes"
	"github.com/cznic/kv"
	"encoding/gob"
	"io"
	"os"
)

var (
	oidKey = []byte("oid")
)

type DBEntry struct {
	object *Object
	data []byte
}

func (dbe *DBEntry) UpdateData() error {
	buf := &bytes.Buffer{}
	enc := gob.NewEncoder(buf)
	err := enc.Encode(dbe.object.data)
	if err != nil {
		return err
	}
	dbe.data = buf.Bytes()
	return nil
}

type Index interface {
	Name() string
	KeySize(o *DBEntry) int
	ValueSize(o *DBEntry) int
	Find(key interface{}) (*DBEntry, error)
	WriteKey(out io.Writer, o *DBEntry) (int, error)
	WriteValue(out io.Writer, o *DBEntry) (int, error)
	ExplainError(err error, writingKey bool) error
}

type DB struct {
	db *kv.DB
	indexes []Index
}

func (db *DB) PutObject(o *Object) (*Object, error) {
	dbe := &DBEntry{object: o}
	if o.Id() == 0 {
		nid, err := db.incOid()
		if err != nil { return nil, err }
		o.SetId(nid)
	}
	err := db.writeToIndexes(dbe)
	return o, err
}

func (db *DB) incOid() (int64, error) {
	nid, err := db.db.Inc(oidKey, 1)
	return nid, err
}

func (db *DB) writeToIndexes(dbe *DBEntry) error {
	for _, v := range db.indexes {
		keySize := v.KeySize(dbe)
		valueSize := v.ValueSize(dbe)
		if keySize <= 0 || valueSize <= 0 {
			continue
		}
		buf := &bytes.Buffer{}
		buf.Grow(keySize + valueSize)
		buf.Reset()
		_, err := v.WriteKey(buf, dbe)
		if err != nil {
			return v.ExplainError(err, true)
		}
		_, err = v.WriteValue(buf, dbe)
		if err != nil {
			return v.ExplainError(err, false)
		}
		data := buf.Bytes()

		err = db.db.Set(data[:keySize], data[keySize:])
		if err != nil {
			return err
		}
	}
	return nil
}

func NewDb(filename string) (*DB, error) {
	kvdb, err := openOrCreate(filename)
	if err != nil {
		return nil, err
	}
	return &DB{
		db: kvdb,
	}, nil
}

func openOrCreate(dbfile string) (*kv.DB, error){
	_, err := os.Stat(dbfile)
	if os.IsNotExist(err) {
		return kv.Create(dbfile, nil)
	} else {
		return kv.Open(dbfile, nil)
	}
}
