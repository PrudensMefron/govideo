// Package audio adapts private audio preview files to the Wails asset server.
package audio

import (
	"net/http"
	"os"
	"strings"
)

type PreviewStore interface {
	OpenPreview(string) (*os.File, error)
}

func Middleware(store PreviewStore) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !strings.HasPrefix(r.URL.Path, "/audio-preview/") {
				next.ServeHTTP(w, r)
				return
			}
			if r.Method != http.MethodGet && r.Method != http.MethodHead {
				w.Header().Set("Allow", "GET, HEAD")
				w.WriteHeader(http.StatusMethodNotAllowed)
				return
			}
			f, err := store.OpenPreview(strings.TrimPrefix(r.URL.Path, "/audio-preview/"))
			if err != nil {
				http.NotFound(w, r)
				return
			}
			defer f.Close()
			info, err := f.Stat()
			if err != nil {
				http.NotFound(w, r)
				return
			}
			w.Header().Set("Content-Type", "audio/wav")
			w.Header().Set("Cache-Control", "no-store")
			w.Header().Set("X-Content-Type-Options", "nosniff")
			http.ServeContent(w, r, "preview.wav", info.ModTime(), f)
		})
	}
}
