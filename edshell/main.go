package main

import (
	"github.com/andrebq/exp/httpfs"
	"log"
	"net/http"
	"path/filepath"
)

func main() {
	dfs, _ := httpfs.NewDiskFile(".")
	handler := &httpfs.Handler{
		Root: dfs,
	}

	http.Handle("/fs/", http.StripPrefix("/fs/", handler))
	http.Handle("/", http.FileServer(http.Dir(filepath.FromSlash("./static"))))

	if err := http.ListenAndServe("localhost:4001", nil); err != nil {
		log.Printf("Error starting server: %v", err)
	}
}
