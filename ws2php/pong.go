package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

type ppRequest struct {
	clid  string
	data  []byte
	done  chan signal
	empty chan signal
}

type pongServer struct {
	// this channel handle the POST
	pingRequest chan *ppRequest
	// this channel handle the GET
	pongRequest chan *ppRequest
}

func NewPongServer() *pongServer {
	srv := &pongServer{
		pingRequest: make(chan *ppRequest, 10),
		pongRequest: make(chan *ppRequest, 10),
	}
	go srv.loop()
	return srv
}

func (s *pongServer) loop() {
	// hold the pending pings, only one for each client
	// but can be changed
	database := make(map[string][]byte)

	for {
		select {
		case ping := <-s.pingRequest:
			database[ping.clid] = ping.data
		case pong := <-s.pongRequest:
			if data, has := database[pong.clid]; has {
				pong.data = data
				pong.done <- signal{}
			} else {
				pong.empty <- signal{}
			}
		}
	}
}

// Store the given data to the given client
func (s *pongServer) Put(clid string, buf []byte) {
	s.pingRequest <- &ppRequest{clid: clid, data: buf}
}

// Return the data for the client or false if there was no data
func (s *pongServer) Get(clid string) ([]byte, bool) {
	req := &ppRequest{clid: clid, done: make(chan signal, 1), empty: make(chan signal, 1)}
	s.pongRequest <- req
	select {
	case <-req.done:
		return req.data, true
	case <-req.empty:
		return nil, false
	}
	panic("not reached")
	return nil, false
}

func pingPongHandler(server *pongServer) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Method == "POST" {
			clid := req.URL.Path
			buf, err := ioutil.ReadAll(req.Body)
			defer req.Body.Close()
			if err != nil {
				// ignore
				return
			}
			server.Put(clid, buf)
		} else if req.Method == "GET" {
			clid := req.URL.Path
			if buf, have := server.Get(clid); have {
				w.Write(buf)
			}
			// wait for 2 seconds
			<-time.After(2 * time.Second)
			// check again
			if buf, have := server.Get(clid); have {
				w.Write(buf)
			} else {
				// give up, send 404
				// TODO: maybe 404 isn't the best choice here
				// but for the moment, it should work
				http.NotFound(w, req)
			}
		} else {
			http.NotFound(w, req)
		}
	})
}

func startPong() {
	srv := NewPongServer()
	http.Handle("/", pingPongHandler(srv))
	err := http.ListenAndServe(*port, nil)
	if err != nil {
		log.Printf("pongserver | Error: %v", err)
	}
}
