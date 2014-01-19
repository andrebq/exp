package main

import (
	"flag"
	"github.com/andrebq/exp/httpfs"
	"github.com/andrebq/gas"
	"log"
	"net/http"
)

var (
	addr    = flag.String("addr", "0.0.0.0:9091", "Address to listen for clients")
	baseDir = flag.String("baseDir", ".", "Base dir to serve the content")
)

func main() {
	staticDir, err := gas.Abs("github.com/andrebq/exp/davd/site")
	if err != nil {
		log.Printf("Error loading static site directory. cause: %v", err)
		return
	}

	m := &httpfs.Mount{BaseDir: *baseDir}
	httpfs := httpfs.NewHttpFS(m, "/fs/")

	http.Handle("/fs/", http.StripPrefix("/fs", httpfs))
	http.Handle("/", http.FileServer(http.Dir(staticDir)))

	log.Printf("Starting davd server at %v", *addr)
	err = http.ListenAndServe(*addr, nil)
	http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Printf("Error opening server: %v", err)
	}
}
