package app

import (
	"context"
	"encoding/json"
	"github.com/PrudensMefron/govideo/internal/core"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"
)

type fakeExtractor struct{}

func (fakeExtractor) Probe(context.Context, core.Source) (core.Media, error) {
	return core.Media{Title: "x"}, nil
}

type blockingDownloader struct {
	current, max atomic.Int32
	release      chan struct{}
}

func (d *blockingDownloader) Download(ctx context.Context, r core.DownloadRequest, emit core.EventHandler) error {
	n := d.current.Add(1)
	defer d.current.Add(-1)
	for {
		m := d.max.Load()
		if n <= m || d.max.CompareAndSwap(m, n) {
			break
		}
	}
	select {
	case <-d.release:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

type fakeConverter struct{}

func (fakeConverter) Convert(context.Context, string, string, core.AudioOptions, func(core.JobProgress)) (string, error) {
	return "out.mp3", nil
}
func TestWorkerLimitAndCancellation(t *testing.T) {
	d := &blockingDownloader{release: make(chan struct{})}
	s := New(fakeExtractor{}, d, fakeConverter{}, "", 2, nil)
	src := core.Source{URL: "https://example.com/v"}
	m := core.Media{Title: "x"}
	for i := 0; i < 3; i++ {
		if _, err := s.StartDownload(src, m, core.OutputVideo, core.VideoQuality{}, core.AudioOptions{}, t.TempDir()); err != nil {
			t.Fatal(err)
		}
	}
	deadline := time.Now().Add(time.Second)
	for d.max.Load() < 2 && time.Now().Before(deadline) {
		time.Sleep(time.Millisecond)
	}
	if d.max.Load() != 2 {
		t.Fatalf("max concurrency=%d", d.max.Load())
	}
	jobs := s.ListJobs(0, 10)
	if _, err := s.Cancel(jobs[0].ID); err != nil {
		t.Fatal(err)
	}
	close(d.release)
}
func TestInterruptedJobRecovery(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "jobs.json")
	now := time.Now()
	jobs := map[core.JobID]*core.Job{"x": {ID: "x", Status: core.JobDownloading, CreatedAt: now, UpdatedAt: now}}
	b, _ := json.Marshal(jobs)
	if err := os.WriteFile(path, b, 0600); err != nil {
		t.Fatal(err)
	}
	s := New(fakeExtractor{}, &blockingDownloader{release: make(chan struct{})}, fakeConverter{}, path, 1, nil)
	got := s.ListJobs(0, 10)[0]
	if got.Status != core.JobFailed || got.Failure == nil || got.Failure.Code != "interrupted" {
		t.Fatalf("%+v", got)
	}
}

func TestPruneUnavailableActivityPersistsRemoval(t *testing.T) {
	dir := t.TempDir()
	history := filepath.Join(dir, "jobs.json")
	file := filepath.Join(dir, "song.mp3")
	if err := os.WriteFile(file, []byte("audio"), 0600); err != nil {
		t.Fatal(err)
	}
	s := New(nil, nil, nil, history, 1, nil)
	s.RecordSavedAudio(file)
	id := s.ListJobs(0, 10)[0].ID
	if err := s.PruneUnavailableArtifacts(id); err == nil {
		t.Fatal("available file must not be removed from history")
	}
	if err := os.Remove(file); err != nil {
		t.Fatal(err)
	}
	jobs := s.ListJobs(0, 10)
	if len(jobs) != 1 || jobs[0].Artifacts[0].Available {
		t.Fatalf("missing file should be marked unavailable: %+v", jobs)
	}
	if err := s.PruneUnavailableArtifacts(id); err != nil {
		t.Fatal(err)
	}
	if got := s.ListJobs(0, 10); len(got) != 0 {
		t.Fatalf("activity was not removed: %+v", got)
	}
	reopened := New(nil, nil, nil, history, 1, nil)
	if got := reopened.ListJobs(0, 10); len(got) != 0 {
		t.Fatalf("removed activity returned after restart: %+v", got)
	}
}

func TestPruneUnavailableArtifactKeepsAvailableOutput(t *testing.T) {
	dir := t.TempDir()
	history := filepath.Join(dir, "jobs.json")
	video := filepath.Join(dir, "video.mp4")
	music := filepath.Join(dir, "music.mp3")
	if err := os.WriteFile(video, []byte("video"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(music, []byte("music"), 0600); err != nil {
		t.Fatal(err)
	}
	now := time.Now()
	id := core.JobID("both")
	seed := map[core.JobID]*core.Job{id: {
		ID: id, Status: core.JobCompleted, CreatedAt: now, UpdatedAt: now,
		Artifacts: []core.Artifact{{Kind: core.OutputVideo, Path: video}, {Kind: core.OutputAudio, Path: music}},
	}}
	b, err := json.Marshal(seed)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(history, b, 0600); err != nil {
		t.Fatal(err)
	}
	s := New(nil, nil, nil, history, 1, nil)
	if err := os.Remove(music); err != nil {
		t.Fatal(err)
	}
	if err := s.PruneUnavailableArtifacts(id); err != nil {
		t.Fatal(err)
	}
	got := s.ListJobs(0, 10)
	if len(got) != 1 || len(got[0].Artifacts) != 1 || got[0].Artifacts[0].Path != video || !got[0].Artifacts[0].Available {
		t.Fatalf("available output was not preserved: %+v", got)
	}
	if _, err := os.Stat(video); err != nil {
		t.Fatalf("pruning must not delete files: %v", err)
	}
	reopened := New(nil, nil, nil, history, 1, nil)
	if got := reopened.ListJobs(0, 10); len(got) != 1 || len(got[0].Artifacts) != 1 {
		t.Fatalf("partial pruning did not persist: %+v", got)
	}
}
