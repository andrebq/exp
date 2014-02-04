package main

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

var (
	dbtemp = filepath.Join(os.TempDir(),
		fmt.Sprintf("%v", time.Now().UnixNano()))
)

type testLog interface {
	Fatalf(fmt string, args ...interface{})
	Errorf(fmt string, args ...interface{})
	Logf(fmt string, args ...interface{})
}

func mkTempDir(t testLog) {
	err := os.MkdirAll(dbtemp, 0644)
	if err != nil {
		t.Fatalf("Error creating the directory: %v", err)
	}
}

func deleteTempDir(t testLog) {
	err := os.RemoveAll(dbtemp)
	if err != nil {
		t.Logf("Error deleteting the dbtemp dir: %v", err)
	}
}

func createDb(t testLog) *DB {
	switch t := t.(type) {
	case *testing.B:
		t.StopTimer()
	}
	mkTempDir(t)
	defer deleteTempDir(t)
	db, err := CreateDB(dbtemp)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	return db
}

func cleanupDb(db *DB, t testLog) {
	err := db.Close()
	if err != nil {
		t.Logf("Error closing db: %v", err)
	}
	deleteTempDir(t)
}

func TestCreateDB(t *testing.T) {
	mkTempDir(t)
	defer deleteTempDir(t)
	db, err := CreateDB(dbtemp)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if db == nil {
		t.Fatalf("db is nil")
	}

	err = db.Close()
	if err != nil {
		t.Errorf("Unexpected error closing: %v", err)
	}
}

func TestCreateKeyword(t *testing.T) {
	db := createDb(t)
	defer cleanupDb(db, t)

	var err error
	key := NewKeyword(":abc")
	err = db.CreateKeyword(key)
	if err != nil {
		t.Errorf("Unexpected error. %v", err)
	}

	if !key.Valid() {
		t.Errorf("Keyword should be valid. %v", key)
	}

	if !key.ValidName() {
		t.Errorf("Keyword should have a valid name. %v", key)
	}

	other := NewKeyword(":abc")
	err = db.CreateKeyword(other)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if !other.Valid() {
		t.Errorf("The other key should be valid too")
	}

	if !other.ValidName() {
		t.Errorf("The other key should have a valid name")
	}

	if other.val != key.val {
		t.Errorf("They should have the save value. But key is %v and other is %v", key.val, other.val)
	}
}

func BenchmarkCreateNewKeyword(b *testing.B) {
	db := createDb(b)
	defer cleanupDb(db, b)

	var err error

	keys := make([]string, 0, b.N)

	for i := 0; i < b.N; i++ {
		keys = append(keys, fmt.Sprintf(":abc/%v", i))
	}

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		key := NewKeyword(keys[i])
		err = db.CreateKeyword(key)
		if err != nil {
			b.Errorf("Error creating keyword. iteration %v", i)
		}
	}
}

func BenchmarkCreateOneFetchMany(b *testing.B) {
	db := createDb(b)
	defer cleanupDb(db, b)

	var err error

	// create one keyword
	key := NewKeyword(":abc")
	err = db.CreateKeyword(key)
	if err != nil {
		b.Fatalf("Cannot query without creating first. %v", err)
	}

	b.StartTimer()
	other := NewKeyword(":abc")
	for i := 0; i < b.N; i++ {
		err = db.CreateKeyword(other)
		if err != nil {
			b.Errorf("Error checking keyword. iteration %v. cause: %v", i, err)
		}
	}
}
