package httpfs

import (
	"log"
	"net/http"
	"strconv"
	"strings"
)

// Handler
type Handler struct {
	Root File
}

type ReadArgs struct {
	DeepRead bool
	Depth int
}

func (h *Handler) fileForPath(req *http.Request, createIfNotFound bool) (File, error) {
	toWalk := h.extractPath(req)
	if createIfNotFound {
		return OpenOrCreate(h.Root, createIfNotFound, toWalk...)
	}
	return Walk(h.Root, toWalk...)
}

func (h *Handler) extractPath(req *http.Request) []string {
	return strings.Split(req.URL.Path, "/")
}

func (h *Handler) readArgs(req *http.Request) (ReadArgs, error) {
	var ra ReadArgs
	var err error
	ra.DeepRead = req.URL.Query().Get("deepread") == "y"
	if ra.DeepRead {
		var depth int64
		depth, err = strconv.ParseInt(req.URL.Query().Get("depth"), 10, 32)
		ra.Depth = int(depth)
	}
	return ra, err
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case "GET":
		h.Read(w, req)
	case "PUT", "POST":
		h.Write(w, req)
	default:
		http.Error(w, "invalid method", http.StatusMethodNotAllowed)
	}
}

func (h *Handler) Write(w http.ResponseWriter, req *http.Request) {
	file, err := h.fileForPath(req, true)
	if err != nil {
		log.Printf("Error walking path file: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	} else {
		err = Truncate(file)
		if err != nil && err != ErrCannotTruncate {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		err = WriteToFile(file, req.Body)
		if err != nil {
			log.Printf("Error saving file: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		} else {
			w.WriteHeader(http.StatusOK)
		}
	}
}

func (h *Handler) Read(w http.ResponseWriter, req *http.Request) {
	file, err := h.fileForPath(req, false)
	if err != nil {
		log.Printf("Error walking path file: %v", err)
		status := http.StatusInternalServerError
		if err == ErrNotFound {
			status = http.StatusNotFound
		}
		http.Error(w, err.Error(), status)
	} else {
		args, err := h.readArgs(req)
		if err != nil {
			log.Printf("Bad request: %v", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if file.Info().Dir {
			if args.DeepRead {
				err = DeepReadTo(w, file, args.Depth)
			} else {
				err = ReadFileTo(w, file)
			}
		} else {
			err = ReadFileTo(w, file)
		}
		if err != nil {
			log.Printf("Error saving file: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
