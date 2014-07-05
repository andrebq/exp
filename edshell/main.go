package main

import (
	"github.com/andrebq/exp/httpfs"
	"github.com/cznic/kv"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

func mustOpenShellDb(name string) *kv.DB {
	var err error
	var db *kv.DB
	_, err = os.Stat(name)
	if os.IsNotExist(err) {
		db, err = kv.Create(name, &kv.Options{VerifyDbBeforeOpen: true})
	} else if err == nil {
		db, err = kv.Open(name, &kv.Options{VerifyDbBeforeOpen: true})
	}
	if err != nil {
		panic(err)
	}
	return db
}

func main() {
	dfs, _ := httpfs.NewDiskFile(".")
	handler := &httpfs.Handler{
		Root: dfs,
	}
	db := &DB{
		DB: mustOpenShellDb("shell.db"),
	}
	http.Handle("/fs/", http.StripPrefix("/fs/", handler))
	http.Handle("/db/", http.StripPrefix("/db/", db))
	http.Handle("/", http.FileServer(http.Dir(filepath.FromSlash("./static"))))

	if err := http.ListenAndServe("localhost:4001", nil); err != nil {
		log.Printf("Error starting server: %v", err)
	}
}
