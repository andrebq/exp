// Uses the martini framework to render the codemirror editor
package main

import (
	"github.com/andrebq/exp/codemirror"
	"github.com/codegangsta/martini"
	"net/http"
)

func main() {
	m := martini.Classic()

	e := &codemirror.Editor{Prefix: "/gocm/"}
	m.Get("/gocm/**", func(w http.ResponseWriter, req *http.Request) {
		e.ServeHTTP(w, req)
	})

	e2 := &codemirror.Editor{Prefix: "/"}

	m.Get("/**", func(w http.ResponseWriter, req *http.Request) {
		e2.ServeHTTP(w, req)
	})

	m.Run()
}
