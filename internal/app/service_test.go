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
