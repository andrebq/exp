package httpfs

import (
	"log"
	"net/http"
	"strings"
)

// Handler
type Handler struct {
	Root File
}

func (h *Handler) fileForPath(req *http.Request) (File, error) {
	toWalk := h.extractPath(req)
	return Walk(h.Root, toWalk...)
}

func (h *Handler) extractPath(req *http.Request) []string {
	return strings.Split(req.URL.Path, "/")
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
	file, err := h.fileForPath(req)
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
	file, err := h.fileForPath(req)
	if err != nil {
		log.Printf("Error walking path file: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	} else {
		if file.Info().Dir {
			err = ReadFileTo(w, file)
		} else {
			err = ReadFileTo(w, file)
		}
		if err != nil {
			log.Printf("Error saving file: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
