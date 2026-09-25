package audio

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

type fixture struct{ path string }

func (f fixture) OpenPreview(id string) (*os.File, error) {
	if id != "private-token" {
		return nil, os.ErrNotExist
	}
	return os.Open(f.path)
}

func TestPreviewRangeAndTokenIsolation(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.wav")
	if err := os.WriteFile(path, []byte("0123456789"), 0600); err != nil {
		t.Fatal(err)
	}
	h := Middleware(fixture{path})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(204) }))
	r := httptest.NewRequest("GET", "/audio-preview/private-token", nil)
	r.Header.Set("Range", "bytes=2-5")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != 206 || w.Body.String() != "2345" || w.Header().Get("Content-Type") != "audio/wav" {
		t.Fatalf("%d %s %v", w.Code, w.Body.String(), w.Header())
	}
	for _, url := range []string{"/audio-preview/unknown", "/audio-preview/../../etc/passwd", "/audio-preview/"} {
		w := httptest.NewRecorder()
		h.ServeHTTP(w, httptest.NewRequest("GET", url, nil))
		if w.Code != 404 {
			t.Fatal(w.Code)
		}
	}
	w = httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("POST", "/audio-preview/private-token", nil))
	if w.Code != 405 {
		t.Fatal(w.Code)
	}
	w = httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", "/", nil))
	if w.Code != 204 {
		t.Fatal(w.Code)
	}
}
