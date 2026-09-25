package app

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/PrudensMefron/govideo/internal/core"
)

// ResolveArtifact permits only completed media outputs, not arbitrary executable paths.
func (s *Service) ResolveArtifact(path string) (string, error) {
	if !filepath.IsAbs(path) {
		return "", fmt.Errorf("caminho de arquivo inválido")
	}
	path = filepath.Clean(path)
	if !core.SupportedAudioPath(path) {
		switch strings.ToLower(filepath.Ext(path)) {
		case ".mp4", ".mkv", ".webm", ".mov", ".avi", ".m4v":
		default:
			return "", fmt.Errorf("este tipo de arquivo não pode ser aberto pelo GoVideo")
		}
	}
	s.mu.RLock()
	known := false
	for _, j := range s.jobs {
		if j.Status != core.JobCompleted {
			continue
		}
		for _, a := range j.Artifacts {
			if filepath.Clean(a.Path) == path {
				known = true
				break
			}
		}
		if known {
			break
		}
	}
	s.mu.RUnlock()
	if !known {
		return "", fmt.Errorf("o arquivo não pertence a uma atividade concluída")
	}
	info, err := os.Lstat(path)
	if err != nil {
		return "", fmt.Errorf("o arquivo foi movido, removido ou está indisponível")
	}
	if !info.Mode().IsRegular() {
		return "", fmt.Errorf("o arquivo não é uma mídia regular")
	}
	return path, nil
}

func (s *Service) ListAudioArtifacts() []core.AudioFile {
	result := make([]core.AudioFile, 0)
	seen := map[string]bool{}
	for _, j := range s.ListJobs(0, 0) {
		if j.Status != core.JobCompleted {
			continue
		}
		for _, a := range j.Artifacts {
			if a.Kind != core.OutputAudio || !core.SupportedAudioPath(a.Path) || seen[a.Path] {
				continue
			}
			info, err := os.Stat(a.Path)
			if err != nil || !info.Mode().IsRegular() {
				continue
			}
			seen[a.Path] = true
			result = append(result, core.AudioFile{Path: a.Path, Name: info.Name(), SizeBytes: info.Size(), Format: strings.TrimPrefix(strings.ToLower(filepath.Ext(a.Path)), ".")})
		}
	}
	return result
}

func (s *Service) RecordSavedAudio(path string) {
	now := time.Now()
	p := 1.0
	j := &core.Job{ID: core.JobID(fmt.Sprintf("job-%d-%d", now.UnixMilli(), s.seq.Add(1))), Kind: core.JobAudioTrim, InputPath: path, Mode: core.OutputAudio, Status: core.JobCompleted, Destination: filepath.Dir(path), Progress: core.JobProgress{Phase: core.PhaseFinalizing, Fraction: &p}, Artifacts: []core.Artifact{{Kind: core.OutputAudio, Path: path, Available: true}}, CreatedAt: now, UpdatedAt: now, StartedAt: &now, FinishedAt: &now}
	s.put(j)
}
