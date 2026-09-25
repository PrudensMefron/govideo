package ffmpeg

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/PrudensMefron/govideo/internal/core"
	"github.com/PrudensMefron/govideo/internal/process"
)

func (a *Adapter) InspectAudio(ctx context.Context, path string) (float64, error) {
	out, err := process.CommandContext(ctx, a.FFprobe, "-v", "error", "-select_streams", "a:0", "-show_entries", "format=duration:stream=duration,codec_type", "-of", "json", path).Output()
	if err != nil {
		return 0, fmt.Errorf("não foi possível analisar o áudio; verifique se o FFprobe está disponível: %w", err)
	}
	var p struct {
		Format struct {
			Duration string `json:"duration"`
		} `json:"format"`
		Streams []struct {
			Duration string `json:"duration"`
		} `json:"streams"`
	}
	if err := json.Unmarshal(out, &p); err != nil {
		return 0, fmt.Errorf("resposta inválida do FFprobe: %w", err)
	}
	if len(p.Streams) == 0 {
		return 0, fmt.Errorf("o arquivo não contém uma faixa de áudio")
	}
	d, _ := strconv.ParseFloat(p.Streams[0].Duration, 64)
	if d <= 0 {
		d, _ = strconv.ParseFloat(p.Format.Duration, 64)
	}
	if d <= 0 || math.IsNaN(d) || math.IsInf(d, 0) {
		return 0, fmt.Errorf("não foi possível determinar a duração do áudio")
	}
	return d, nil
}

func trimArgs(input, output string, start, duration float64) ([]string, error) {
	if !core.SupportedAudioPath(output) {
		return nil, fmt.Errorf("formato de áudio não suportado")
	}
	args := []string{"-nostdin", "-v", "error", "-n", "-i", input, "-ss", strconv.FormatFloat(start, 'f', 6, 64), "-t", strconv.FormatFloat(duration, 'f', 6, 64), "-map", "0:a:0", "-vn", "-map_metadata", "0", "-map_chapters", "-1"}
	switch strings.ToLower(filepath.Ext(output)) {
	case ".mp3":
		args = append(args, "-c:a", "libmp3lame", "-q:a", "2")
	case ".m4a":
		args = append(args, "-c:a", "aac", "-b:a", "192k", "-movflags", "+faststart")
	case ".aac":
		args = append(args, "-c:a", "aac", "-b:a", "192k")
	case ".opus":
		args = append(args, "-c:a", "libopus", "-b:a", "160k")
	case ".ogg":
		args = append(args, "-c:a", "libvorbis", "-q:a", "6")
	case ".flac":
		args = append(args, "-c:a", "flac")
	case ".wav":
		args = append(args, "-c:a", "pcm_s16le")
	}
	return append(args, output), nil
}

// TrimAudio decodes before trimming, so cuts need not fall on compressed packet boundaries.
func (a *Adapter) TrimAudio(ctx context.Context, input, output string, start, duration float64) error {
	if err := requireNewOutput(output); err != nil {
		return err
	}
	args, err := trimArgs(input, output, start, duration)
	if err != nil {
		return err
	}
	if err := process.CommandContext(ctx, a.FFmpeg, args...).Run(); err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		return fmt.Errorf("não foi possível recortar o áudio; verifique o FFmpeg e o espaço em disco: %w", err)
	}
	return nil
}

// PreviewWAV decodes the final encoded artifact for consistent WebView playback.
func (a *Adapter) PreviewWAV(ctx context.Context, input, output string) error {
	if err := requireNewOutput(output); err != nil {
		return err
	}
	err := process.CommandContext(ctx, a.FFmpeg, "-nostdin", "-v", "error", "-n", "-i", input, "-map", "0:a:0", "-vn", "-c:a", "pcm_s16le", "-ar", "48000", "-ac", "2", output).Run()
	if err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		return fmt.Errorf("não foi possível preparar a reprodução da prévia: %w", err)
	}
	return nil
}

func requireNewOutput(path string) error {
	if _, err := os.Lstat(path); err == nil {
		return fmt.Errorf("o arquivo de saída já existe")
	} else if !os.IsNotExist(err) {
		return err
	}
	return nil
}
