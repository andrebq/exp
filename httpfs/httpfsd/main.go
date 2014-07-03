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
	fs, err := httpfs.NewDiskFile(*dir)
	if err != nil {
		log.Fatalf("error loading root file: %v", err)
	}
	http.Handle("/", &httpfs.Handler{fs})

	log.Printf("starting server at: %v", *addr)

	if err := http.ListenAndServe(*addr, nil); err != nil {
		log.Fatalf("error: %v", err)
	}
}
