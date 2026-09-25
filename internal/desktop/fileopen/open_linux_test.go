package fileopen

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDefaultOpenerReceivesOneLiteralArgument(t *testing.T) {
	dir := t.TempDir()
	received := filepath.Join(dir, "received")
	t.Setenv("GOVIDEO_OPEN_TEST", received)
	if err := os.WriteFile(filepath.Join(dir, "xdg-open"), []byte("#!/bin/sh\nprintf '%s' \"$1\" > \"$GOVIDEO_OPEN_TEST\"\n"), 0700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir)
	path := filepath.Join(dir, "música & $(echo unsafe); 'test'.mp3")
	if err := Open(path); err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if b, err := os.ReadFile(received); err == nil && string(b) == path {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("default application did not receive the exact file path")
}

func TestMissingDefaultOpenerReportsError(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	if err := Open("/tmp/music.mp3"); err == nil {
		t.Fatal("missing system handler was not reported")
	}
}

func TestAssociationFailureReportsError(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "xdg-open"), []byte("#!/bin/sh\nexit 3\n"), 0700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir)
	if err := Open("/tmp/music.mp3"); err == nil {
		t.Fatal("association failure was not reported")
	}
}
