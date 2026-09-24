package deps

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type roundTrip func(*http.Request) (*http.Response, error)

func (f roundTrip) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
func TestInstallChecksumAndAtomicTarget(t *testing.T) {
	body := []byte("binary")
	sum := sha256.Sum256(body)
	m := New(t.TempDir(), nil)
	m.client = &http.Client{Transport: roundTrip(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 200, Status: "200 OK", Body: io.NopCloser(strings.NewReader(string(body))), ContentLength: int64(len(body))}, nil
	})}
	target := filepath.Join(m.root, "bin", "tool")
	if err := m.install(context.Background(), Asset{"tool", "https://example.test/tool", fmt.Sprintf("%x", sum), 0755}, target, 100); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(target); err != nil {
		t.Fatal(err)
	}
	if err := m.install(context.Background(), Asset{"tool", "https://example.test/tool", fmt.Sprintf("%064d", 0), 0755}, target+"2", 100); err == nil {
		t.Fatal("expected checksum error")
	}
	if _, err := os.Stat(target + "2"); !os.IsNotExist(err) {
		t.Fatal("bad asset became visible")
	}
}
