package main

import (
	"github.com/cznic/bufs"
	"github.com/cznic/kv"
	"io/ioutil"
	"net/http"
	"strconv"
)

type DB struct {
	*kv.DB
	cache bufs.Cache
}

func (db *DB) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case "POST", "PUT":
		db.insert(w, req)
	case "GET":
		db.retreive(w, req)
	}
}

func (db *DB) keyFromRequest(req *http.Request) string {
	return req.URL.Path
}

func (db *DB) insert(w http.ResponseWriter, req *http.Request) {
	key := db.keyFromRequest(req)
	data, err := ioutil.ReadAll(req.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = db.Set([]byte(key), data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (db *DB) retreive(w http.ResponseWriter, req *http.Request) {
	key := db.keyFromRequest(req)
	cached := db.cache.Cget(4096)
	defer db.cache.Put(cached)
	data, err := db.Get(cached, []byte(key))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if data == nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Length", strconv.FormatInt(int64(len(data)), 10))
	w.Write(data)
}
