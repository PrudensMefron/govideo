package ffmpeg

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/PrudensMefron/govideo/internal/core"
	"github.com/PrudensMefron/govideo/internal/platformfs"
	"github.com/PrudensMefron/govideo/internal/process"
)

type Adapter struct{ FFmpeg, FFprobe string }

func New(ffmpeg, ffprobe string) *Adapter {
	if ffmpeg == "" {
		ffmpeg = "ffmpeg"
	}
	if ffprobe == "" {
		ffprobe = "ffprobe"
	}
	return &Adapter{ffmpeg, ffprobe}
}

type probe struct {
	Format struct {
		Duration string `json:"duration"`
	} `json:"format"`
	Streams []struct {
		CodecType string `json:"codec_type"`
	} `json:"streams"`
}

func (a *Adapter) Inspect(ctx context.Context, path string) (float64, error) {
	out, err := process.CommandContext(ctx, a.FFprobe, "-v", "error", "-show_entries", "format=duration:stream=codec_type", "-of", "json", path).Output()
	if err != nil {
		return 0, fmt.Errorf("ffprobe: %w", err)
	}
	var p probe
	if err = json.Unmarshal(out, &p); err != nil {
		return 0, err
	}
	ok := false
	for _, s := range p.Streams {
		if s.CodecType == "video" {
			ok = true
		}
	}
	if !ok {
		return 0, fmt.Errorf("input has no video stream")
	}
	d, _ := strconv.ParseFloat(p.Format.Duration, 64)
	return d, nil
}
func (a *Adapter) Convert(ctx context.Context, input, dest string, opt core.AudioOptions, progress func(core.JobProgress)) (string, error) {
	b, err := core.AudioBitrate(opt.Format, opt.Quality)
	if err != nil {
		return "", err
	}
	d, err := a.Inspect(ctx, input)
	if err != nil {
		return "", err
	}
	out, err := platformfs.AvailablePath(dest, strings.TrimSuffix(filepath.Base(input), filepath.Ext(input)), string(opt.Format))
	if err != nil {
		return "", err
	}
	args := []string{"-nostdin", "-i", input, "-vn", "-b:a", fmt.Sprintf("%dk", b), "-progress", "pipe:1", "-nostats", out}
	cmd := process.CommandContext(ctx, a.FFmpeg, args...)
	pipe, err := cmd.StdoutPipe()
	if err != nil {
		return "", err
	}
	if err = cmd.Start(); err != nil {
		return "", err
	}
	s := bufio.NewScanner(pipe)
	for s.Scan() {
		parts := strings.SplitN(s.Text(), "=", 2)
		if len(parts) == 2 && parts[0] == "out_time_ms" && progress != nil {
			us, _ := strconv.ParseFloat(parts[1], 64)
			f := us / 1e6 / d
			if f > 1 {
				f = 1
			}
			progress(core.JobProgress{Phase: core.PhaseTranscoding, Fraction: &f})
		}
	}
	if err = cmd.Wait(); err != nil {
		return "", fmt.Errorf("ffmpeg: %w", err)
	}
	return out, nil
}
