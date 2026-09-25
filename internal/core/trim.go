package core

import (
	"fmt"
	"math"
	"path/filepath"
	"strings"
)

// TrimRequest removes seconds from either edge, never from an absolute end timestamp.
type TrimRequest struct {
	Path         string  `json:"path"`
	StartSeconds float64 `json:"startSeconds"`
	EndSeconds   float64 `json:"endSeconds"`
}

type AudioFile struct {
	Path            string  `json:"path"`
	Name            string  `json:"name"`
	DurationSeconds float64 `json:"durationSeconds"`
	SizeBytes       int64   `json:"sizeBytes"`
	Format          string  `json:"format"`
}

func TrimDuration(duration, start, end float64) (float64, error) {
	for _, n := range []float64{duration, start, end} {
		if math.IsNaN(n) || math.IsInf(n, 0) || n < 0 {
			return 0, fmt.Errorf("informe tempos válidos e não negativos")
		}
	}
	if start+end == 0 {
		return 0, fmt.Errorf("informe um corte no início ou no final")
	}
	if duration-start-end < 0.1 {
		return 0, fmt.Errorf("os cortes devem deixar pelo menos 0,1 segundo de áudio")
	}
	return duration - start - end, nil
}

func SupportedAudioPath(path string) bool {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".mp3", ".m4a", ".opus", ".wav", ".flac", ".ogg", ".aac":
		return true
	}
	return false
}
