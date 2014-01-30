package main

import (
	"fmt"
	"testing"
	"os"
	"path/filepath"
	"time"
)

var (
	dbtemp = filepath.Join(os.TempDir(),
		fmt.Sprintf("%v", time.Now().UnixNano()))
)

func mkTempDir(t *testing.T) {
	err := os.MkdirAll(dbtemp, 0644)
	if err != nil {
		t.Fatalf("Error creating the directory: %v", err)
	}
}

func deleteTempDir(t *testing.T) {
	err := os.Remove(dbtemp)
	if err != nil {
		t.Logf("Error deleteting the dbtemp dir: %v", err)
	}
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
