// CodeMirror wraps a simple text editor based on CodeMirror.
//
// Users of this package must provide a prefix from which all
// codemirror files will be loaded and a diretory with the conteents
// that should be used as the editor.
package codemirror

import (
	"bytes"
	"fmt"
	"github.com/andrebq/gas"
	"html/template"
	"io"
	"log"
	"net/http"
	"net/url"
	"path"
	"path/filepath"
	"strings"
)

var (
	ids    = make(chan int, 0)
	assets = gas.MustAbs("github.com/andrebq/exp/codemirror/assets")
)

func generateIds() {
	id := int(0)
	for {
		id++
		ids <- id
	}
}

func init() {
	go generateIds()
}

type Editor struct {
	// hold extra dependencies that should be loaded before codemirror
	deps []*url.URL
	// hold extra scritps that should be inserted after codemirror is loaded
	modules []*url.URL
	// hold all styles that should be loaded after codemirror.css
	styles []*url.URL

	// prefix to remove from the incoming url and to
	// append for output urls
	Prefix string
}

// ServeHTTP handles serving the text editor contents.
//
// Only static files are served from here
func (e *Editor) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	oldUrl, newUrl := e.stripPrefix(req)
	println("oldURL, newURL", oldUrl.String(), newUrl.String())
	if strings.HasPrefix(req.URL.Path, "single-editor.html") {
		// must return the html render of the template
		req.URL = oldUrl
		e.renderEditor(w, req, false)
		req.URL = newUrl
	} else if strings.HasPrefix(req.URL.Path, "editor-part.html") {
		req.URL = oldUrl
		e.renderEditor(w, req, true)
		req.URL = newUrl
	} else {
		// a static file, delegate to http.ServeFile
		http.ServeFile(w, req, filepath.Join(assets, filepath.FromSlash(req.URL.Path)))
	}
}

func (e *Editor) renderEditor(w http.ResponseWriter, req *http.Request, onlyPart bool) {
	tmp := &bytes.Buffer{}
	tmplName := "full"
	if onlyPart {
		tmplName = "partial"
	}
	codemirrorTmpl := template.Must(template.New("").ParseFiles(gas.MustAbs("github.com/andrebq/exp/codemirror/codemirror.html")))

	err := codemirrorTmpl.ExecuteTemplate(tmp, tmplName, map[string]interface{}{
		"editorid": fmt.Sprintf("%v", <-ids),
		"libcss": []string{
			path.Join(req.URL.Path, "..", "lib/codemirror.css"),
			path.Join(req.URL.Path, "..", "lib/gocm.css"),
		},
		"libscript": []string{
			path.Join(req.URL.Path, "..", "lib/codemirror.js"),
			path.Join(req.URL.Path, "..", "lib/autosize.js"),
		},
	})
	if err != nil {
		log.Printf("[codemirror] error rendering template: %v", err)
		http.Error(w, "error", http.StatusInternalServerError)
		return
	}
	io.Copy(w, tmp)
}

func (e *Editor) stripPrefix(req *http.Request) (full *url.URL, striped *url.URL) {
	if strings.HasPrefix(req.URL.Path, e.Prefix) {
		// copy the old url just in case
		old := *req.URL
		req.URL.Path = req.URL.Path[len(e.Prefix):]
		return &old, req.URL
	}
	// no changes, just return the same value
	return req.URL, req.URL
}

// HandlerFunc wraps the editor under the http.HandlerFunc interface
func (e *Editor) HandlerFunc() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		println("handler func")
		e.ServeHTTP(w, req)
	})
}
