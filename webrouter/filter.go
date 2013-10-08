package webrouter

import (
	"net/http"
	"sync"
)

// Represent a chain of filters
type Filter struct {
	sync.RWMutex
	before  []http.Handler
	handler http.Handler
	after   []http.Handler
}

// Create a empty filter that always return 404
func NewFilter() *Filter {
	return &Filter{before: make([]http.Handler, 0),
		handler: http.NotFoundHandler(),
		after:   make([]http.Handler, 0)}
}

// Inlucde h as a handler called before the actual handler is
// executed
func (f *Filter) PushBefore(h http.Handler) {
	f.Lock()
	defer f.Unlock()
	f.before = append(f.before, h)
}

// Include h as a handler called after the actual handler is
// executed
func (f *Filter) PushAfter(h http.Handler) {
	f.Lock()
	defer f.Unlock()
	f.after = append(f.after, h)
}

// Set the actual handler, if a nil value is passed,
// the Filter will return 404
func (f *Filter) SetHandler(h http.Handler) {
	f.Lock()
	defer f.Unlock()
	f.handler = h
}

func silentPanic() { _ = recover() }

func safeCallHandler(h http.Handler, w http.ResponseWriter, req *http.Request) {
	// prevent any panic from propagating
	defer silentPanic()
	h.ServeHTTP(w, req)
}

// Implement the http.Handler interface
func (f *Filter) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	f.RLock()
	defer f.RUnlock()
	for _, v := range f.before {
		safeCallHandler(v, w, req)
	}
	safeCallHandler(f.handler, w, req)
	for _, v := range f.after {
		safeCallHandler(v, w, req)
	}
}
