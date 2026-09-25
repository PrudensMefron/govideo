package audio

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"
)

func TestLoopbackPreviewPlaybackAndIsolation(t *testing.T) {
	path := filepath.Join(t.TempDir(), "audio.wav")
	if err := os.WriteFile(path, []byte("0123456789"), 0600); err != nil {
		t.Fatal(err)
	}
	s, err := NewServer(fixture{path})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(s.Close)
	req, _ := http.NewRequest("GET", s.URL("private-token"), nil)
	req.Header.Set("Range", "bytes=2-5")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	b, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil || resp.StatusCode != 206 || string(b) != "2345" {
		t.Fatalf("response %d %q %v", resp.StatusCode, b, err)
	}
	req, _ = http.NewRequest("GET", s.URL("private-token"), nil)
	req.Host = "attacker.example"
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != 403 {
		t.Fatalf("foreign host accepted: %d", resp.StatusCode)
	}
	resp, err = http.Get(s.URL("unknown"))
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != 404 {
		t.Fatalf("unknown token accepted: %d", resp.StatusCode)
	}
	s.Close()
	if resp, err := http.Get(s.URL("private-token")); err == nil {
		resp.Body.Close()
		t.Fatal("listener survived shutdown")
	}
}
