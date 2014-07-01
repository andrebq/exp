package main

import (
	"flag"
	"github.com/andrebq/exp/pandora"
	pandorahttp "github.com/andrebq/exp/pandora/http"
	"github.com/andrebq/exp/pandora/pgstore"
	"github.com/andrebq/exp/pandora/webui"
	"log"
	"net/http"
)

var (
	addr   = flag.String("addr", "0.0.0.0:4003", "Address to listen for incoming requests. Used also to serve the webui")
	static = flag.String("staticDir", "!usegas", "Static directory to serve webui files. By default uses gas (requires that a valid GOPATH is set")
	h      = flag.Bool("h", false, "Help")
)

func main() {
	flag.Parse()
	if *h {
		flag.Usage()
		return
	}

	messageStore, err := pgstore.OpenMessageStore("pandora", "pandora", "localhost", "pandora")
	if err != nil {
		log.Fatalf("error opening message store: %v", err)
	}
	blobStore, err := pgstore.OpenBlobStore("pandora", "pandora", "localhost", "pandora")
	if err != nil {
		log.Fatalf("error opening blob store: %v", err)
	}

	server := pandora.Server{
		BlobStore:    blobStore,
		MessageStore: messageStore,
	}

	handler := &webui.Handler{
		Api: pandorahttp.Handler{
			Server:     &server,
			AllowAdmin: true,
		},
	}

	if *static == "!usegas" {
		handler.DefaultStatic()
	} else {
		handler.Static = http.FileServer(http.Dir(*static))
	}

	log.Printf("starting server at %v", *addr)
	if err = http.ListenAndServe(*addr, handler); err != nil {
		log.Fatalf("error starting server: %v", err)
	}
}
