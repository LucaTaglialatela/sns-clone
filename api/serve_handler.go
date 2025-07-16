package api

import (
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/HENNGE/snsclone-202506-golang-luca/frontend"
)

type ServeHandler struct{
	Fs http.Handler
}

func NewServeHandler(fs http.Handler) *ServeHandler {
	return &ServeHandler{
		Fs: fs,
	}
}

func (h *ServeHandler) Serve(w http.ResponseWriter, r *http.Request) {
	// This SPA routing logic works perfectly with the new setup.
	checkPath := strings.TrimPrefix(r.URL.Path, "/")
	if checkPath == "" {
		checkPath = "index.html"
	}

	file, err := frontend.DistFS.Open(checkPath)
	if err != nil {
		// If file doesn't exist, serve index.html
		index, err := frontend.DistFS.Open("index.html")
		if err != nil {
			http.NotFound(w, r)
			return
		}
		defer index.Close()
		http.ServeContent(w, r, "index.html", time.Now(), index.(io.ReadSeeker))
		return
	}
	file.Close()

	// Let the file server handle the existing file.
	h.Fs.ServeHTTP(w, r)
}
