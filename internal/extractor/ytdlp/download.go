package ytdlp

import (
	"bufio"
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/PrudensMefron/govideo/internal/core"
	"github.com/PrudensMefron/govideo/internal/process"
)

func (a *Adapter) Download(ctx context.Context, request core.DownloadRequest, emit core.EventHandler) error {
	args, err := DownloadArgs(request)
	if err != nil {
		return err
	}
	cmd := process.CommandContext(ctx, a.binary, args...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	var stderr strings.Builder
	cmd.Stderr = &stderr
	if err = cmd.Start(); err != nil {
		return fmt.Errorf("start yt-dlp: %w", err)
	}
	if emit != nil {
		emit(core.DownloadStarted{JobID: request.JobID})
	}
	s := bufio.NewScanner(stdout)
	for s.Scan() {
		line := s.Text()
		if e, ok := parseProgress(line, request.JobID); ok && emit != nil {
			emit(e)
		}
		if strings.HasPrefix(line, "GOVIDEO_PATH|") && emit != nil {
			emit(core.DownloadCompleted{JobID: request.JobID, Path: strings.TrimPrefix(line, "GOVIDEO_PATH|")})
		}
	}
	if err = s.Err(); err != nil {
		return err
	}
	if err = cmd.Wait(); err != nil {
		return &CommandError{Err: err, Stderr: stderr.String()}
	}
	return nil
}
func DownloadArgs(r core.DownloadRequest) ([]string, error) {
	if err := r.Mode.Validate(); err != nil {
		return nil, err
	}
	title := "%(title).180B"
	output := r.Destination + "/" + title + ".%(ext)s"
	args := []string{"--no-playlist", "--newline", "--progress-template", "download:GOVIDEO|%(progress.downloaded_bytes)s|%(progress.total_bytes_estimate)s|%(progress.speed)s|%(progress.eta)s", "--print", "after_move:GOVIDEO_PATH|%(filepath)s", "-P", r.Destination, "-o", output}
	h := r.Video.MaxHeight
	video := "bv*[ext=mp4]+ba[ext=m4a]/b[ext=mp4]"
	if h > 0 {
		video = fmt.Sprintf("bv*[height<=%d][ext=mp4]+ba[ext=m4a]/b[height<=%d][ext=mp4]", h, h)
	}
	switch r.Mode {
	case core.OutputVideo:
		args = append(args, "-f", video, "--merge-output-format", "mp4")
	case core.OutputAudio:
		b, e := core.AudioBitrate(r.Audio.Format, r.Audio.Quality)
		if e != nil {
			return nil, e
		}
		args = append(args, "-x", "--audio-format", string(r.Audio.Format), "--audio-quality", fmt.Sprintf("%dk", b))
	case core.OutputBoth:
		return nil, fmt.Errorf("%w: both is orchestrated as two safe invocations", core.ErrInvalidOption)
	}
	args = appendHeaders(args, r.Source.Headers)
	return append(args, r.Source.URL), nil
}
func parseProgress(line string, id core.JobID) (core.DownloadProgress, bool) {
	p := strings.Split(strings.TrimSpace(line), "|")
	if len(p) != 5 || p[0] != "GOVIDEO" {
		return core.DownloadProgress{}, false
	}
	num := func(s string) float64 { v, _ := strconv.ParseFloat(s, 64); return v }
	return core.DownloadProgress{JobID: id, DownloadedBytes: int64(num(p[1])), TotalBytes: int64(num(p[2])), SpeedBytesPerSecond: num(p[3]), ETA: 0}, true
}
