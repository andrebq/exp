package main

import (
	"flag"
	"github.com/andrebq/exp/httpfs"
	"log"
	"net/http"
)

var (
	addr    = flag.String("addr", "0.0.0.0:9091", "Address to listen for clients")
	baseDir = flag.String("baseDir", ".", "Base dir to serve the content")
)

func index(w http.ResponseWriter, req *http.Request) {
}

func main() {

	m := &httpfs.Mount{BaseDir: *baseDir}
	httpfs := httpfs.NewHttpFS(m, "/fs/")

	http.Handle("/fs/", http.StripPrefix("/fs", httpfs))
	http.HandleFunc("/", index)

	log.Printf("Starting davd server at %v", *addr)
	err := http.ListenAndServe(*addr, nil)
	http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Printf("Error opening server: %v", err)
	}
}
