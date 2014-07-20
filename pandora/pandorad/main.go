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
	dbUser = flag.String("dbUser", "pandora", "User to access the pandora database")
	dbPasswd = flag.String("dbPasswd", "pandora", "Password to access the pandora database")
	dbHost = flag.String("dbHost", "localhost", "Host to connect")
	dbName = flag.String("dbName", "pandpra", "Name of the database to connect")

	initPgStore = flag.Bool("initPgStore", false, "Initialize the tables on the database")
	h      = flag.Bool("h", false, "Help")
)

func main() {
	flag.Parse()
	if *h {
		flag.Usage()
		return
	}

	messageStore, err := pgstore.OpenMessageStore(*dbUser, *dbPasswd, *dbHost, *dbName)
	if err != nil {
		log.Fatalf("error opening message store: %v", err)
	}
	blobStore, err := pgstore.OpenBlobStore(*dbUser, *dbPasswd, *dbHost, *dbName)
	if err != nil {
		log.Fatalf("error opening blob store: %v", err)
	}

	if *initPgStore {
		if err := messageStore.InitTables(); err != nil {
			log.Fatalf("Error initializing message tables: %v", err)
		}

		if err := blobStore.InitTables(); err != nil {
			log.Fatalf("Error initializing blob tables: %v", err)
		}

		log.Printf("Tables initialized")
		return
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
