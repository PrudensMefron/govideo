package app

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/PrudensMefron/govideo/internal/core"
)

type AudioEngine interface {
	InspectAudio(context.Context, string) (float64, error)
	TrimAudio(context.Context, string, string, float64, float64) error
	PreviewWAV(context.Context, string, string) error
}

type AudioPreview struct {
	ID              string         `json:"id"`
	Source          core.AudioFile `json:"source"`
	StartSeconds    float64        `json:"startSeconds"`
	EndSeconds      float64        `json:"endSeconds"`
	DurationSeconds float64        `json:"durationSeconds"`
}

type SaveAudioRequest struct {
	PreviewID string `json:"previewID"`
	Overwrite bool   `json:"overwrite"`
}

type SavedAudio struct {
	Path        string `json:"path"`
	Overwritten bool   `json:"overwritten"`
}

type audioPreviewData struct {
	AudioPreview
	directory, output, playback string
	sourceInfo                  os.FileInfo
	digest                      [sha256.Size]byte
}

// AudioEditor owns one temporary preview per application, with bounded/cancellable work.
// Its tokens grant access only to prepared playback files, never arbitrary filesystem paths.
type AudioEditor struct {
	engine  AudioEngine
	root    string
	work    sync.Mutex
	mu      sync.Mutex
	preview *audioPreviewData
	ctx     context.Context
	cancel  context.CancelFunc
}

func NewAudioEditor(engine AudioEngine) (*AudioEditor, error) {
	root, err := os.MkdirTemp("", "govideo-audio-")
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithCancel(context.Background())
	return &AudioEditor{engine: engine, root: root, ctx: ctx, cancel: cancel}, nil
}

func (s *AudioEditor) Close() {
	s.cancel()
	s.work.Lock()
	defer s.work.Unlock()
	s.mu.Lock()
	s.preview = nil
	s.mu.Unlock()
	_ = os.RemoveAll(s.root) // Only the private directory allocated by MkdirTemp.
}

func audioSource(path string) (os.FileInfo, error) {
	if !filepath.IsAbs(path) || !core.SupportedAudioPath(path) {
		return nil, fmt.Errorf("selecione um arquivo MP3, M4A, Opus, WAV, FLAC, Ogg ou AAC")
	}
	info, err := os.Lstat(path)
	if err != nil {
		return nil, fmt.Errorf("o arquivo não está disponível: %w", err)
	}
	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("selecione um arquivo de áudio regular, não uma pasta ou atalho")
	}
	return info, nil
}

func (s *AudioEditor) Inspect(ctx context.Context, path string) (core.AudioFile, error) {
	info, err := audioSource(path)
	if err != nil {
		return core.AudioFile{}, err
	}
	duration, err := s.engine.InspectAudio(ctx, path)
	if err != nil {
		return core.AudioFile{}, err
	}
	return core.AudioFile{Path: path, Name: info.Name(), DurationSeconds: duration, SizeBytes: info.Size(), Format: strings.TrimPrefix(strings.ToLower(filepath.Ext(path)), ".")}, nil
}

func (s *AudioEditor) CreatePreview(ctx context.Context, r core.TrimRequest) (AudioPreview, error) {
	if !s.work.TryLock() {
		return AudioPreview{}, fmt.Errorf("aguarde a operação de áudio em andamento")
	}
	defer s.work.Unlock()
	ctx, cancel := context.WithCancel(ctx)
	stop := context.AfterFunc(s.ctx, cancel)
	defer stop()
	defer cancel()
	if s.ctx.Err() != nil {
		return AudioPreview{}, context.Canceled
	}
	source, err := s.Inspect(ctx, r.Path)
	if err != nil {
		return AudioPreview{}, err
	}
	duration, err := core.TrimDuration(source.DurationSeconds, r.StartSeconds, r.EndSeconds)
	if err != nil {
		return AudioPreview{}, err
	}
	info, err := audioSource(r.Path)
	if err != nil {
		return AudioPreview{}, err
	}
	digest, err := audioDigest(ctx, r.Path)
	if err != nil {
		return AudioPreview{}, err
	}
	dir, err := os.MkdirTemp(s.root, "preview-")
	if err != nil {
		return AudioPreview{}, err
	}
	keep := false
	defer func() {
		if !keep {
			_ = os.RemoveAll(dir)
		}
	}()
	output := filepath.Join(dir, "trimmed."+source.Format)
	if err = s.engine.TrimAudio(ctx, r.Path, output, r.StartSeconds, duration); err != nil {
		return AudioPreview{}, err
	}
	playback := filepath.Join(dir, "playback.wav")
	if err = s.engine.PreviewWAV(ctx, output, playback); err != nil {
		return AudioPreview{}, err
	}
	latest, err := audioDigest(ctx, r.Path)
	if err != nil {
		return AudioPreview{}, err
	}
	if digest != latest {
		return AudioPreview{}, fmt.Errorf("o arquivo original mudou; selecione-o novamente")
	}
	var token [16]byte
	if _, err = rand.Read(token[:]); err != nil {
		return AudioPreview{}, err
	}
	p := &audioPreviewData{AudioPreview: AudioPreview{ID: hex.EncodeToString(token[:]), Source: source, StartSeconds: r.StartSeconds, EndSeconds: r.EndSeconds, DurationSeconds: duration}, directory: dir, output: output, playback: playback, sourceInfo: info, digest: digest}
	s.mu.Lock()
	old := s.preview
	s.preview = p
	s.mu.Unlock()
	keep = true
	if old != nil {
		_ = os.RemoveAll(old.directory)
	}
	return p.AudioPreview, nil
}

func (s *AudioEditor) OpenPreview(id string) (*os.File, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.preview == nil || s.preview.ID != id {
		return nil, os.ErrNotExist
	}
	return os.Open(s.preview.playback)
}

func (s *AudioEditor) Discard(id string) error {
	s.work.Lock()
	defer s.work.Unlock()
	s.mu.Lock()
	p := s.preview
	if p == nil || p.ID != id {
		s.mu.Unlock()
		return nil
	}
	s.preview = nil
	s.mu.Unlock()
	return os.RemoveAll(p.directory)
}

func (s *AudioEditor) Save(ctx context.Context, r SaveAudioRequest) (SavedAudio, error) {
	if !s.work.TryLock() {
		return SavedAudio{}, fmt.Errorf("aguarde a operação de áudio em andamento")
	}
	defer s.work.Unlock()
	s.mu.Lock()
	p := s.preview
	s.mu.Unlock()
	if p == nil || p.ID != r.PreviewID {
		return SavedAudio{}, fmt.Errorf("gere uma nova prévia antes de salvar")
	}
	if err := ctx.Err(); err != nil {
		return SavedAudio{}, err
	}
	var path string
	var err error
	if r.Overwrite {
		path = p.Source.Path
		// Stage the complete artifact beside the original before replacing it.
		var staged *os.File
		staged, err = os.CreateTemp(filepath.Dir(path), ".govideo-trim-*")
		if err != nil {
			return SavedAudio{}, err
		}
		defer os.Remove(staged.Name())
		err = copyAudio(p.output, staged, p.sourceInfo.Mode().Perm())
		if err != nil {
			return SavedAudio{}, err
		}
		info, err := audioSource(path)
		if err != nil {
			return SavedAudio{}, err
		}
		digest, err := audioDigest(ctx, path)
		if err != nil {
			return SavedAudio{}, err
		}
		if !os.SameFile(info, p.sourceInfo) || digest != p.digest {
			return SavedAudio{}, fmt.Errorf("o original mudou desde a prévia; gere outra prévia ou salve uma cópia")
		}
		err = os.Rename(staged.Name(), path)
	} else {
		path, err = saveAudioCopy(p)
	}
	if err != nil {
		return SavedAudio{}, fmt.Errorf("não foi possível salvar o áudio: %w", err)
	}
	// Consume only after successful publication; a failed save is retryable.
	s.mu.Lock()
	s.preview = nil
	s.mu.Unlock()
	_ = os.RemoveAll(p.directory)
	return SavedAudio{Path: path, Overwritten: r.Overwrite}, nil
}

func saveAudioCopy(p *audioPreviewData) (string, error) {
	ext := filepath.Ext(p.Source.Path)
	base := strings.TrimSuffix(p.Source.Name, ext) + " - recortada"
	for n := 0; n < 10000; n++ {
		suffix := ""
		if n > 0 {
			suffix = fmt.Sprintf(" (%d)", n)
		}
		path := filepath.Join(filepath.Dir(p.Source.Path), base+suffix+ext)
		out, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
		if os.IsExist(err) {
			continue
		}
		if err != nil {
			return "", err
		}
		if err := copyAudio(p.output, out, p.sourceInfo.Mode().Perm()); err != nil {
			_ = os.Remove(path)
			return "", err
		}
		return path, nil
	}
	return "", fmt.Errorf("não foi possível escolher um nome livre")
}

func copyAudio(source string, out *os.File, mode os.FileMode) (err error) {
	defer func() {
		if e := out.Close(); err == nil {
			err = e
		}
	}()
	in, err := os.Open(source)
	if err != nil {
		return err
	}
	defer in.Close()
	if _, err = io.Copy(out, in); err != nil {
		return err
	}
	if err = out.Chmod(mode); err != nil {
		return err
	}
	return out.Sync()
}

func audioDigest(ctx context.Context, path string) ([sha256.Size]byte, error) {
	var digest [sha256.Size]byte
	f, err := os.Open(path)
	if err != nil {
		return digest, err
	}
	defer f.Close()
	h := sha256.New()
	buf := make([]byte, 128*1024)
	for {
		if err := ctx.Err(); err != nil {
			return digest, err
		}
		n, err := f.Read(buf)
		if n > 0 {
			_, _ = h.Write(buf[:n])
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return digest, err
		}
	}
	copy(digest[:], h.Sum(nil))
	return digest, nil
}
