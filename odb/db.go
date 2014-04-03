package odb

import (
	"bytes"
	"github.com/cznic/kv"
	"os"
)

type DBEntry struct {
	*Object
	data []byte
}

func (dbe *DBEntry) InvalidateData() []byte {
	ret := dbe.data
	dbe.data = nil
	return ret
}

func (dbe *DBEntry) UpdateData() ([]byte, error) {
	if dbe.data != nil {
		return dbe.data, nil
	}

	buf := &bytes.Buffer{}
	bw := &BinaryWriter{buf, nil}
	err := bw.WriteTypedMap(&dbe.TypedMap)
	if err != nil {
		return nil, err
	}
	dbe.data = buf.Bytes()
	return dbe.data, nil
}

type Index interface {
	Name() string
	Find(values ...interface{}) (*DBEntry, error)
	Write(dbe *DBEntry) error
	ExplainError(err error, writingKey bool) error
}

type DB struct {
	db      *kv.DB
	indexes []Index
}

func (db *DB) PutObject(o *Object) (*Object, error) {
	dbe := &DBEntry{Object: o, data: nil}
	err := db.writeToIndexes(dbe)
	return o, err
}

func (db *DB) FindByOID(vals ...interface{}) (*Object, error) {
	return db.FindOneByIndex("core_oid", vals...)
}

func (db *DB) FindOneByIndex(idxName string, vals ...interface{}) (*Object, error) {
	for _, v := range db.indexes {
		if v.Name() == idxName {
			dbe, err := v.Find(vals...)
			return dbe.Object, err
		}
	}
	return nil, errNoIndexProvided
}

func (db *DB) writeToIndexes(dbe *DBEntry) error {
	var err error
	for _, v := range db.indexes {
		err = v.Write(dbe)
		if err != nil {
			return v.ExplainError(err, false)
		}
	}
	return nil
}

func NewDB(filename string) (*DB, error) {
	kvdb, err := openOrCreate(filename)
	if err != nil {
		return nil, err
	}
	db := &DB{
		db:      kvdb,
		indexes: make([]Index, 0),
	}
	db.AddIndex(&OidIndex{kvdb})
	return db, nil
}

func (db *DB) AddIndex(idx Index) {
	db.indexes = append(db.indexes, idx)
}

func openOrCreate(dbfile string) (*kv.DB, error) {
	opt := &kv.Options{VerifyDbBeforeOpen: true,
		VerifyDbAfterOpen:   true,
		VerifyDbBeforeClose: true,
		VerifyDbAfterClose:  true}

	if len(dbfile) == 0 {
		// in memory database
		return kv.CreateMem(opt)
	}
	_, err := os.Stat(dbfile)
	if os.IsNotExist(err) {
		return kv.Create(dbfile, opt)
	} else {
		return kv.Open(dbfile, opt)
	}
}
