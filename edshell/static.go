package main

import (
	"net/http"
)

type ServeStaticFiles struct {
	staticServer http.Handler
}

func (s *ServeStaticFiles) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	s.staticServer.ServeHTTP(w, req)
}
