package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
)

type Mount struct {
	baseDir string
}

func (m *Mount) realPath(u *url.URL) string {
	return filepath.Join(m.baseDir, filepath.FromSlash(u.Path))
}

func (m *Mount) InfoFromURL(u *url.URL) (os.FileInfo, error) {
	return os.Stat(m.realPath(u))
}

func (m *Mount) OpenReadFile(u *url.URL) (*os.File, error) {
	return os.Open(m.realPath(u))
}

func (m *Mount) ReadDir(u *url.URL) ([]os.FileInfo, error) {
	f, err := os.Open(m.realPath(u))
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return f.Readdir(-1)
}

func (m *Mount) CreateOrOpenFileForWrite(u *url.URL) (*os.File, error) {
	rp := m.realPath(u)
	stat, err := os.Stat(rp)
	if os.IsNotExist(err) {
		// new file
		d := filepath.Dir(rp)
		err = os.MkdirAll(d, 0644)
		if err != nil {
			return nil, err
		}

		return os.OpenFile(rp, os.O_CREATE, 0644)
	}
	if err != nil {
		return nil, err
	}
	if stat.IsDir() {
		return nil, fmt.Errorf("%v is a directory", u.Path)
	}
	return os.OpenFile(rp, os.O_RDWR, 0644)
}

type RawFS struct {
	mount    *Mount
	metaBase string
}

func (r *RawFS) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	info, err := r.mount.InfoFromURL(req.URL)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	if info.IsDir() {
		r.serveDir(w, req, info)
	} else {
		r.serveFile(w, req, info)
	}
}

func (r *RawFS) serveDir(w http.ResponseWriter, req *http.Request, info os.FileInfo) {
	childs, err := r.mount.ReadDir(req.URL)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	fmt.Fprintf(w,
		`<!doctype html>
<html>
<head>
	<title>Listing directory: %v</title>
</head>
<body>
	<h1>Listing directory: %v</h1>
	<ul>`,
		info.Name(),
		info.Name())

	for _, child := range childs {
		fmt.Fprintf(w, `<li><a href="%v">%v</a> <a href="%v">Stat</a></li>`,
			"./"+child.Name(), child.Name(), r.metaForChild(child.Name(), req.URL))
	}

	fmt.Fprintf(w,
		`	</ul>
</body>
</html>`)
}

func (r *RawFS) metaForChild(name string, parent *url.URL) *url.URL {
	u := parent.ResolveReference(&url.URL{})
	u.Path = path.Join(r.metaBase, u.Path, name)
	return u
}

func (r *RawFS) serveFile(w http.ResponseWriter, req *http.Request, info os.FileInfo) {
	switch req.Method {
	case "GET":
		r.serveGetFile(w, req, info)
	case "POST", "PUT":
		r.servePostFile(w, req, info)
	default:
		if req.Method != "GET" {
			http.Error(w, "Only GET/POST/PUT at this moment", http.StatusMethodNotAllowed)
			return
		}
	}
}

func (r *RawFS) servePostFile(w http.ResponseWriter, req *http.Request, info os.FileInfo) {
	http.Error(w, "Not implemented yet!", 500)
	return
	file, err := r.mount.CreateOrOpenFileForWrite(req.URL)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer file.Close()
	_, err = io.Copy(file, req.Body)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
}

func (r *RawFS) serveGetFile(w http.ResponseWriter, req *http.Request, info os.FileInfo) {
	rw, err := r.mount.OpenReadFile(req.URL)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer rw.Close()

	http.ServeContent(w, req, info.Name(), info.ModTime(), rw)
}

type MetaFS struct {
	mount   *Mount
	rawBase string
}

func (m *MetaFS) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	info, err := m.mount.InfoFromURL(req.URL)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	m.printAsHtml(w, req, info)
}

// Return the raw url representing this file
func (m *MetaFS) rawURLFor(meta *url.URL) *url.URL {
	u := meta.ResolveReference(&url.URL{})
	u.Path = path.Join(m.rawBase, meta.Path)
	return u
}

func (m *MetaFS) printAsHtml(w http.ResponseWriter, req *http.Request, info os.FileInfo) {
	rawUrl := m.rawURLFor(req.URL)

	fmt.Fprintf(w,
		`<!doctype html>
<head>
	<title>Info about: %v</title>
</head>
<body>
	<dl>
		<dt>Name</dt> <dd>%v</dd>
		<dt>Directory?</dt> <dd>%v</dd>
		<dt>Size</dt> <dd>%d</dd>
		<dt>Mod time</dt> <dd>%v</dd>
		<dt>Raw url</td> <dd><a href="%v" rel="nofollow">%v</href></dd>
	</dl>
</body>`,
		req.URL.Path,
		info.Name(),
		info.IsDir(),
		info.Size(),
		info.ModTime(),
		rawUrl,
		info.Name())
}

var (
	addr    = flag.String("addr", "0.0.0.0:9091", "Address to listen for clients")
	baseDir = flag.String("baseDir", ".", "Base dir to serve the content")
)

func index(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w,
		`<!doctype html>
<head>
	<title>Root dir</title>
</head>
<body>
	<dl>
		<dt>Meta</dt> <dd><a href="./meta/">./meta/</a></dd>
		<del><dt>Raw</dt> <dd>./raw/</dd></del>
	</dl>
</body>`)
}

func main() {

	m := &Mount{baseDir: *baseDir}
	metafs := &MetaFS{mount: m, rawBase: "/raw"}
	rawfs := &RawFS{mount: m, metaBase: "/meta"}

	http.HandleFunc("/", index)
	http.Handle("/meta/", http.StripPrefix("/meta/", metafs))
	http.Handle("/raw/", http.StripPrefix("/raw/", rawfs))

	log.Printf("Starting davd server at %v", *addr)
	err := http.ListenAndServe(*addr, nil)
	http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Printf("Error opening server: %v", err)
	}
}
