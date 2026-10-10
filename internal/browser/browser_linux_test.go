//go:build linux

package browser

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestSystemBrowserLinux(t *testing.T) {
	dir := t.TempDir()
	binary := filepath.Join(dir, "xdg-mime")
	// Isolate detection from the user's installed browsers and cookie stores.
	t.Setenv("PATH", dir)
	for _, tt := range []struct{ output, want string }{
		{"org.mozilla.firefox.desktop", "firefox"},
		{"com.google.Chrome.desktop", "chrome"},
		{"unsupported.desktop", ""},
		{"", ""},
	} {
		script := "#!/bin/sh\n[ \"$1 $2 $3 $4\" = 'query default x-scheme-handler/https ' ] || exit 2\nprintf '%s' '" + tt.output + "'\n"
		if err := os.WriteFile(binary, []byte(script), 0700); err != nil {
			t.Fatal(err)
		}
		got, err := Resolve(context.Background(), "system")
		if got != tt.want || (err == nil) != (tt.want != "") {
			t.Fatalf("%q: got %q, %v", tt.output, got, err)
		}
	}
	if err := os.Remove(binary); err != nil {
		t.Fatal(err)
	}
	if _, err := Resolve(context.Background(), ""); err == nil {
		t.Fatal("missing detector must produce an actionable error")
	}
}
