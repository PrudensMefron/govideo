//go:build linux

package ytdlp

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/PrudensMefron/govideo/internal/core"
)

func TestCookiesPassedToActualCommands(t *testing.T) {
	dir := t.TempDir()
	binary := filepath.Join(dir, "fake-ytdlp")
	argv := filepath.Join(dir, "argv")
	fixture, err := filepath.Abs("testdata/info.json")
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("GOVIDEO_TEST_ARGV", argv)
	t.Setenv("GOVIDEO_TEST_FIXTURE", fixture)
	script := "#!/bin/sh\nprintf '%s\\n' \"$@\" > \"$GOVIDEO_TEST_ARGV\"\ncase \" $* \" in\n  *' --dump-single-json '*) cat \"$GOVIDEO_TEST_FIXTURE\" ;;\n  *) printf 'GOVIDEO|50|100|25|2\\nGOVIDEO_PATH|/tmp/out.mp3\\n' ;;\nesac\n"
	if err := os.WriteFile(binary, []byte(script), 0700); err != nil {
		t.Fatal(err)
	}
	a := New(binary)
	selected := "firefox"
	a.SetCookieBrowser(func(context.Context) (string, error) { return selected, nil })
	src := core.Source{URL: "https://example.com/video"}
	if _, err := a.Probe(context.Background(), src); err != nil {
		t.Fatal(err)
	}
	assertArgs := func(browser string) {
		t.Helper()
		data, err := os.ReadFile(argv)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.HasPrefix(string(data), "--cookies-from-browser\n"+browser+"\n") {
			t.Fatalf("missing cookie arguments: %s", data)
		}
	}
	assertArgs("firefox")
	selected = "chrome"
	for _, mode := range []core.OutputMode{core.OutputVideo, core.OutputAudio} {
		r := core.DownloadRequest{Source: src, Mode: mode, Destination: dir, Audio: core.AudioOptions{Format: core.AudioMP3, Quality: core.AudioStandard}}
		if err := a.Download(context.Background(), r, nil); err != nil {
			t.Fatal(err)
		}
		assertArgs("chrome")
	}
}
