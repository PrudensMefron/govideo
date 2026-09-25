package app

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/PrudensMefron/govideo/internal/core"
)

type fakeAudio struct {
	trim func(context.Context, string, string, float64, float64) error
}

func (f fakeAudio) InspectAudio(context.Context, string) (float64, error) { return 240, nil }
func (f fakeAudio) TrimAudio(ctx context.Context, in, out string, start, duration float64) error {
	if f.trim != nil {
		return f.trim(ctx, in, out, start, duration)
	}
	return os.WriteFile(out, []byte("trimmed audio"), 0600)
}
func (f fakeAudio) PreviewWAV(_ context.Context, in, out string) error {
	b, err := os.ReadFile(in)
	if err != nil {
		return err
	}
	return os.WriteFile(out, b, 0600)
}

func testAudioEditor(t *testing.T, engine AudioEngine) (*AudioEditor, string) {
	t.Helper()
	s, err := NewAudioEditor(engine)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(s.Close)
	path := filepath.Join(t.TempDir(), "Música & teste.mp3")
	if err := os.WriteFile(path, []byte("original"), 0644); err != nil {
		t.Fatal(err)
	}
	return s, path
}

func previewFor(t *testing.T, s *AudioEditor, path string) AudioPreview {
	t.Helper()
	p, err := s.CreatePreview(context.Background(), core.TrimRequest{Path: path, StartSeconds: 20, EndSeconds: 15})
	if err != nil {
		t.Fatal(err)
	}
	if p.DurationSeconds != 205 {
		t.Fatalf("duration %v", p.DurationSeconds)
	}
	return p
}

func TestAudioPreviewCopyAndOverwrite(t *testing.T) {
	s, path := testAudioEditor(t, fakeAudio{})
	ctx := context.Background()
	p := previewFor(t, s, path)
	if b, _ := os.ReadFile(path); string(b) != "original" {
		t.Fatal("preview changed original")
	}
	if _, err := s.OpenPreview("../../" + path); !errors.Is(err, os.ErrNotExist) {
		t.Fatal("accepted arbitrary path")
	}
	f, err := s.OpenPreview(p.ID)
	if err != nil {
		t.Fatal(err)
	}
	f.Close()
	copy1, err := s.Save(ctx, SaveAudioRequest{PreviewID: p.ID})
	if err != nil {
		t.Fatal(err)
	}
	if b, _ := os.ReadFile(path); string(b) != "original" {
		t.Fatal("copy changed original")
	}
	if b, _ := os.ReadFile(copy1.Path); string(b) != "trimmed audio" {
		t.Fatal("copy mismatch")
	}
	if _, err := s.OpenPreview(p.ID); err == nil {
		t.Fatal("consumed preview still accessible")
	}
	if _, err := s.Save(ctx, SaveAudioRequest{PreviewID: p.ID}); err == nil {
		t.Fatal("saved a stale preview")
	}
	p = previewFor(t, s, path)
	copy2, err := s.Save(ctx, SaveAudioRequest{PreviewID: p.ID})
	if err != nil {
		t.Fatal(err)
	}
	if copy1.Path == copy2.Path || !strings.Contains(copy2.Path, " (1)") {
		t.Fatal("collision not handled")
	}
	p = previewFor(t, s, path)
	saved, err := s.Save(ctx, SaveAudioRequest{PreviewID: p.ID, Overwrite: true})
	if err != nil {
		t.Fatal(err)
	}
	if saved.Path != path || !saved.Overwritten {
		t.Fatal(saved)
	}
	if b, _ := os.ReadFile(path); string(b) != "trimmed audio" {
		t.Fatal("replacement mismatch")
	}
}

func TestChangedOriginalPreventsOverwriteButAllowsCopy(t *testing.T) {
	s, path := testAudioEditor(t, fakeAudio{})
	p := previewFor(t, s, path)
	if err := os.WriteFile(path, []byte("modified externally"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Save(context.Background(), SaveAudioRequest{PreviewID: p.ID, Overwrite: true}); err == nil {
		t.Fatal("overwrote changed original")
	}
	if b, _ := os.ReadFile(path); string(b) != "modified externally" {
		t.Fatal("original lost")
	}
	if _, err := s.Save(context.Background(), SaveAudioRequest{PreviewID: p.ID}); err != nil {
		t.Fatal(err)
	}
}

func TestPreviewCancellationBoundsConcurrencyAndCleansFiles(t *testing.T) {
	started := make(chan struct{})
	s, path := testAudioEditor(t, fakeAudio{trim: func(ctx context.Context, _, out string, _, _ float64) error {
		if err := os.WriteFile(out, []byte("partial"), 0600); err != nil {
			return err
		}
		close(started)
		<-ctx.Done()
		return ctx.Err()
	}})
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { _, err := s.CreatePreview(ctx, core.TrimRequest{Path: path, StartSeconds: 1}); done <- err }()
	<-started
	if _, err := s.CreatePreview(context.Background(), core.TrimRequest{Path: path, StartSeconds: 1}); err == nil {
		t.Fatal("concurrent preview accepted")
	}
	cancel()
	if err := <-done; !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	entries, err := os.ReadDir(s.root)
	if err != nil || len(entries) != 0 {
		t.Fatalf("partial files remain: %v, %v", entries, err)
	}
	if b, _ := os.ReadFile(path); string(b) != "original" {
		t.Fatal("original lost")
	}
}

func TestFailedPreviewPreservesOriginalAndPreviousPreview(t *testing.T) {
	fail := false
	s, path := testAudioEditor(t, fakeAudio{trim: func(_ context.Context, _, out string, _, _ float64) error {
		if fail {
			return errors.New("encoder failure")
		}
		return os.WriteFile(out, []byte("trimmed"), 0600)
	}})
	p := previewFor(t, s, path)
	fail = true
	if _, err := s.CreatePreview(context.Background(), core.TrimRequest{Path: path, StartSeconds: 1}); err == nil {
		t.Fatal("missing error")
	}
	f, err := s.OpenPreview(p.ID)
	if err != nil {
		t.Fatal(err)
	}
	f.Close()
	if err := s.Discard(p.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := s.OpenPreview(p.ID); err == nil {
		t.Fatal("discard did not revoke token")
	}
}

func TestAudioSourceRejectsNonAudioAndMissingFile(t *testing.T) {
	s, path := testAudioEditor(t, fakeAudio{})
	for _, path := range []string{filepath.Dir(path), path + ".exe", "https://example.com/a.mp3", filepath.Join(filepath.Dir(path), "missing.mp3")} {
		if _, err := s.Inspect(context.Background(), path); err == nil {
			t.Errorf("accepted %s", path)
		}
	}
}

func TestSavedAudioIsListedAndOnlyKnownMediaCanOpen(t *testing.T) {
	s := New(nil, nil, nil, filepath.Join(t.TempDir(), "jobs.json"), 1, nil)
	path := filepath.Join(t.TempDir(), "track.mp3")
	if err := os.WriteFile(path, []byte("audio"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := s.ResolveArtifact(path); err == nil {
		t.Fatal("unknown path accepted")
	}
	s.RecordSavedAudio(path)
	if got, err := s.ResolveArtifact(path); err != nil || got != path {
		t.Fatalf("%s %v", got, err)
	}
	if got := s.ListAudioArtifacts(); len(got) != 1 || got[0].Path != path {
		t.Fatal(got)
	}
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	if _, err := s.ResolveArtifact(path); err == nil {
		t.Fatal("missing output accepted")
	}
	if len(s.ListAudioArtifacts()) != 0 {
		t.Fatal("missing output listed")
	}
}
