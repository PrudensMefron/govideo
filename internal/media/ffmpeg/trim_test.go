package ffmpeg

import (
	"bytes"
	"context"
	"encoding/binary"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestRealFFmpegTrimKeepsExactInterval(t *testing.T) {
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		t.Skip("ffmpeg not installed")
	}
	if _, err := exec.LookPath("ffprobe"); err != nil {
		t.Skip("ffprobe not installed")
	}
	dir := t.TempDir()
	source := filepath.Join(dir, "music with spaces.wav")
	const rate = 48000
	pcm := make([]byte, 5*rate*2)
	for i := 0; i < 5*rate; i++ {
		binary.LittleEndian.PutUint16(pcm[i*2:], uint16((i/rate+1)*1000))
	}
	var wav bytes.Buffer
	wav.WriteString("RIFF")
	binary.Write(&wav, binary.LittleEndian, uint32(36+len(pcm)))
	wav.WriteString("WAVEfmt ")
	for _, v := range []any{uint32(16), uint16(1), uint16(1), uint32(rate), uint32(rate * 2), uint16(2), uint16(16)} {
		binary.Write(&wav, binary.LittleEndian, v)
	}
	wav.WriteString("data")
	binary.Write(&wav, binary.LittleEndian, uint32(len(pcm)))
	wav.Write(pcm)
	if err := os.WriteFile(source, wav.Bytes(), 0600); err != nil {
		t.Fatal(err)
	}
	a := New("", "")
	ctx := context.Background()
	d, err := a.InspectAudio(ctx, source)
	if err != nil || d != 5 {
		t.Fatalf("duration=%v err=%v", d, err)
	}
	out := filepath.Join(dir, "trim.wav")
	if err := a.TrimAudio(ctx, source, out, 1.25, d-1.25-1.5); err != nil {
		t.Fatal(err)
	}
	decoded, err := exec.Command("ffmpeg", "-v", "error", "-i", out, "-f", "s16le", "-acodec", "pcm_s16le", "pipe:1").Output()
	if err != nil {
		t.Fatal(err)
	}
	want := pcm[int(1.25*rate)*2 : int(3.5*rate)*2]
	if !bytes.Equal(decoded, want) {
		t.Fatalf("cut differs from source interval [1.25,3.5): got %d bytes, want %d", len(decoded), len(want))
	}
	for _, ext := range []string{"mp3", "m4a", "opus", "flac", "ogg", "aac"} {
		t.Run(ext, func(t *testing.T) {
			out := filepath.Join(dir, "trim."+ext)
			if err := a.TrimAudio(ctx, source, out, 1.25, 2.25); err != nil {
				t.Fatal(err)
			}
			if duration, err := a.InspectAudio(ctx, out); err != nil || math.Abs(duration-2.25) > .15 {
				t.Fatalf("duration=%v err=%v", duration, err)
			}
			if err := a.PreviewWAV(ctx, out, out+".wav"); err != nil {
				t.Fatal(err)
			}
		})
	}
	// A preview never overwrites an existing destination, even accidentally.
	if err := a.TrimAudio(ctx, source, out, 1, 1); err == nil {
		t.Fatal("existing output overwritten")
	}
}
