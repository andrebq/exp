package main

import (
	"flag"
	"github.com/andrebq/exp/httpfs"
	"log"
	"net/http"
)

var (
	dir  = flag.String("dir", ".", "Directory to expose over httpfs")
	help = flag.Bool("h", false, "Show this menu")
	addr = flag.String("addr", "localhost:4001", "Address to listen for incomming connections")
)

func main() {
	flag.Parse()

	if *help {
		flag.Usage()
		return
	}
	mount := &httpfs.Mount{
		BaseDir: *dir,
	}

	log.Printf("Exposing directory: %v under address %v", *dir, *addr)

	fs := httpfs.NewHttpFS(mount, "/fs")
	http.Handle("/fs/", http.StripPrefix("/fs", fs))

	if err := http.ListenAndServe(*addr, nil); err != nil {
		log.Fatalf("error: %v", err)
	}
}
