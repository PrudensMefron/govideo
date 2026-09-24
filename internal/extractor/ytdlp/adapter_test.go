package ytdlp

import (
	"encoding/json"
	"github.com/PrudensMefron/govideo/internal/core"
	"os"
	"strings"
	"testing"
)

func TestMapFixture(t *testing.T) {
	b, err := os.ReadFile("testdata/info.json")
	if err != nil {
		t.Fatal(err)
	}
	var raw rawInfo
	if err = json.Unmarshal(b, &raw); err != nil {
		t.Fatal(err)
	}
	m := mapMedia(raw)
	if m.Title != "Exemplo" || m.SizeBytes != 1500 || !m.SizeEstimated || len(m.Formats) != 1 || m.Formats[0].AudioCodec != "mp4a" {
		t.Fatalf("unexpected media: %+v", m)
	}
}
func TestDownloadArgs(t *testing.T) {
	base := core.DownloadRequest{Source: core.Source{URL: "https://example.com/v"}, Destination: "/tmp/out", Audio: core.AudioOptions{Format: core.AudioMP3, Quality: core.AudioHigh}, Video: core.VideoQuality{MaxHeight: 1080}}
	for _, mode := range []core.OutputMode{core.OutputVideo, core.OutputAudio} {
		base.Mode = mode
		a, err := DownloadArgs(base)
		if err != nil {
			t.Fatal(err)
		}
		joined := strings.Join(a, " ")
		if !strings.Contains(joined, "https://example.com/v") || !strings.Contains(joined, "--progress-template") {
			t.Fatal(joined)
		}
	}
	base.Mode = core.OutputBoth
	if _, err := DownloadArgs(base); err == nil {
		t.Fatal("both must be orchestrated")
	}
}
func TestProgress(t *testing.T) {
	e, ok := parseProgress("GOVIDEO|50|100|25|2", "x")
	if !ok || e.DownloadedBytes != 50 || e.TotalBytes != 100 || e.SpeedBytesPerSecond != 25 {
		t.Fatalf("%+v %v", e, ok)
	}
}
