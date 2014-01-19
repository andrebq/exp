package httpfs 

import (
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

	if req.Method == "POST" {
		r.serveFile(w, req, info)
		return
	}

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
<html itemscope>
<head>
	<title>Listing directory: %v</title>
</head>
<body>
	<h1>Listing directory: <data itemprop="name">%v</data></h1>
	<ul>`,
		info.Name(),
		info.Name())

	for _, child := range childs {
		fmt.Fprintf(w, `<li itemscope itemprop="child"><a itemprop="url" href="%v"><span itemprop="name">%v</span></a> Directory? <span itemprop="dir">%v</span> / <a itemprop="metaurl" href="%v">Stat</a></li>`,
			"./"+child.Name(), child.Name(), child.IsDir(), r.metaForChild(child.Name(), req.URL))
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

type IndexFS struct {
	prefix string
}

func (i *IndexFS) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w,
		`<!doctype html>
<head>
	<title>Root dir</title>
</head>
<body>
	<dl>
		<dt>Meta</dt> <dd><a href="./meta/">meta/</a></dd>
		<dt>Raw</dt> <dd><a href="./raw/">raw/</a></dd>
	</dl>
</body>`)
}

type HttpFS struct {
	meta *MetaFS
	raw  *RawFS
	idx  *IndexFS
	mux  *http.ServeMux
}

func NewHttpFS(m *Mount, prefix string) *HttpFS {
	r := &HttpFS{}
	prefix = path.Clean(prefix)
	r.raw = &RawFS{mount: m, metaBase: path.Join(prefix, "/meta/")}
	r.meta = &MetaFS{mount: m, rawBase: path.Join(prefix, "/raw/")}
	r.idx = &IndexFS{prefix: prefix}
	r.mux = http.NewServeMux()
	r.mux.Handle("/meta/", http.StripPrefix("/meta", r.meta))
	r.mux.Handle("/raw/", http.StripPrefix("/raw", r.raw))
	r.mux.Handle("/", r.idx)
	return r
}

func (h *HttpFS) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	log.Printf("[httpfs]-[%v]-%v", req.Method, req.URL)
	h.mux.ServeHTTP(w, req)
}
