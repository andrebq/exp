package main

import (
	"flag"
	"io"
	"log"
	"net/http"
	"os"
)

var (
	httpMethod = flag.String("method", "GET", "Http METHOD to use")
	help       = flag.Bool("h", false, "Help menu")
)

func main() {
	flag.Parse()
	if *help {
		flag.Usage()
		return
	}

	if len(flag.Args()) < 1 {
		log.Printf("No target defined")
		return
	}

	remoteUrl := flag.Args()[0]

	log.Printf("Remote address: %v", remoteUrl)
	log.Printf("HTTP Method: %v", *httpMethod)

	switch *httpMethod {
	case "GET":
		handleGet(remoteUrl)
	case "POST":
		handlePost(remoteUrl)
	default:
		log.Printf("Invalid method")
	}
}

func handleGet(addr string) {
	resp, err := http.Get(addr)
	if err != nil {
		log.Printf("Error: %v", err)
	}
	defer resp.Body.Close()
	io.Copy(os.Stdout, resp.Body)
}

func handlePost(addr string) {
	resp, err := http.Post(addr, "application/octet-stream", os.Stdin)
	if err != nil {
		log.Printf("Error: %v", err)
	}
	defer resp.Body.Close()
	io.Copy(os.Stdout, resp.Body)
}
