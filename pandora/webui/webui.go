package webui

import (
	pandorahttp "github.com/andrebq/exp/pandora/http"
	"github.com/andrebq/gas"
	"net/http"
	"strings"
)

type Handler struct {
	Api    pandorahttp.PandoraHandler
	Static http.Handler
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if strings.Index(req.URL.Path, "/api/") > 0 {
		h.Api.ServeHTTP(w, req)
	} else {
		h.Static.ServeHTTP(w, req)
	}
}

func (h *Handler) DefaultStatic() *Handler {
	h.Static = http.FileServer(http.Dir(gas.MustAbs("github.com/andrebq/exp/pandora/webui/static")))
	return h
}
