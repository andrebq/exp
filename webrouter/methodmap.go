package webrouter

import (
	"net/http"
)

// A handler that map http request to specific handlers
type MethodMap struct {
	Post   http.Handler
	Get    http.Handler
	Put    http.Handler
	Delete http.Handler
	Head   http.Handler
	Others http.Handler
}

func NewCrudMap(create http.HandlerFunc, retreive http.HandlerFunc, update http.HandlerFunc, delete http.HandlerFunc) *MethodMap {
	mm := &MethodMap{
		Get:    retreive,
		Post:   create,
		Put:    update,
		Delete: delete,
		Head:   retreive,
		Others: http.NotFoundHandler(),
	}
	return mm
}

// use this to have a nice fallback to all methods
func (mm *MethodMap) callHandler(h http.Handler, w http.ResponseWriter, req *http.Request) {
	if h == nil {
		if mm.Others != nil {
			// try others
			mm.Others.ServeHTTP(w, req)
		} else if mm.Get != nil {
			// if others isn't defined
			// try get
			mm.Get.ServeHTTP(w, req)
		} else {
			// give up
			http.NotFound(w, req)
		}
	} else {
		h.ServeHTTP(w, req)
	}
}

// Implement the net/http interface
func (mm *MethodMap) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case "GET":
		mm.callHandler(mm.Get, w, req)
	case "POST":
		mm.callHandler(mm.Post, w, req)
	case "PUT":
		mm.callHandler(mm.Put, w, req)
	case "DELETE":
		mm.callHandler(mm.Delete, w, req)
	case "HEAD":
		mm.callHandler(mm.Head, w, req)
	default:
		mm.callHandler(mm.Others, w, req)
	}
}
