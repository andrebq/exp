package main

import (
	"github.com/andrebq/exp/httpfs"
	"github.com/cznic/kv"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
)

func mustOpenShellDb(name string) *kv.DB {
	var err error
	var db *kv.DB
	_, err = os.Stat(name)
	if os.IsNotExist(err) {
		db, err = kv.Create(name, &kv.Options{})
	} else if err == nil {
		db, err = kv.Open(name, &kv.Options{})
	}
	if err != nil {
		panic(err)
	}
	return db
}

func cleanup(db *DB) {
	buf := make(chan os.Signal, 1)
	signal.Notify(buf, os.Interrupt, os.Kill)
	// wait for one of the signals
	log.Printf("going out with signal: %v", <-buf)
	// cleanup the database
	if err := db.DB.Rollback(); err != nil {
		log.Printf("error rollingback transactions: %v", err)
	}
	if err := db.DB.BeginTransaction(); err != nil {
		log.Printf("error starting the close transaction: %v", err)
	}
	if err := db.DB.Close(); err != nil {
		log.Printf("error closing the database: %v", err)
	}
	os.Exit(0)
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
	http.Handle("/", &ServeStaticFiles{
		staticServer: http.FileServer(http.Dir(filepath.FromSlash("./static"))),
	})

	go cleanup(db)

	if err := http.ListenAndServe("localhost:4001", nil); err != nil {
		log.Printf("Error starting server: %v", err)
	}
}
