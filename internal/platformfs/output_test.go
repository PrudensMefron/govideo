package platformfs

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAvailablePathDoesNotOverwrite(t *testing.T) {
	d := t.TempDir()
	if err := os.WriteFile(filepath.Join(d, "a.mp3"), []byte("x"), 0600); err != nil {
		t.Fatal(err)
	}
	p, err := AvailablePath(d, "a", "mp3")
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Base(p) != "a (1).mp3" {
		t.Fatal(p)
	}
}
