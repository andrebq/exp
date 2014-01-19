package main

import (
	"bytes"
	"io"
	"io/ioutil"
	"log"
	"net/http"
)

func echoServer(w http.ResponseWriter, req *http.Request) {
	log.Printf("Got %v/%v", req.Method, req.URL)
	buf, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Printf("Error reading body. %v", err)
		return
	}
	err = req.ParseForm()
	if err != nil {
		log.Printf("Error parsing form. %v", err)
	}
	log.Printf("Body contents: %v", string(buf))
	io.Copy(w, bytes.NewBuffer(buf))
}

func startEchoServer(port *string) {
	http.HandleFunc("/", echoServer)
	log.Printf("Starting server at: %v", *port)
	err := http.ListenAndServe(*port, nil)
	if err != nil {
		log.Printf("Error: %v", err)
	}
}
