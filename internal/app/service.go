package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"github.com/PrudensMefron/govideo/internal/core"
)

type Converter interface {
	Convert(context.Context, string, string, core.AudioOptions, func(core.JobProgress)) (string, error)
}
type Service struct {
	mu         sync.RWMutex
	jobs       map[core.JobID]*core.Job
	cancels    map[core.JobID]context.CancelFunc
	queue      chan core.JobID
	extractor  core.Extractor
	downloader core.Downloader
	converter  Converter
	history    string
	notify     func(core.Job)
	seq        atomic.Uint64
}

func New(ex core.Extractor, dl core.Downloader, cv Converter, history string, workers int, notify func(core.Job)) *Service {
	if workers < 1 {
		workers = 2
	}
	s := &Service{jobs: map[core.JobID]*core.Job{}, cancels: map[core.JobID]context.CancelFunc{}, queue: make(chan core.JobID, 128), extractor: ex, downloader: dl, converter: cv, history: history, notify: notify}
	s.load()
	for i := 0; i < workers; i++ {
		go s.worker()
	}
	return s
}
func (s *Service) AnalyzeURL(ctx context.Context, raw string) (core.Media, error) {
	src := core.Source{URL: raw}
	if err := src.Validate(); err != nil {
		return core.Media{}, err
	}
	return s.extractor.Probe(ctx, src)
}
func (s *Service) StartDownload(src core.Source, media core.Media, mode core.OutputMode, video core.VideoQuality, audio core.AudioOptions, dest string) (core.Job, error) {
	if err := src.Validate(); err != nil {
		return core.Job{}, err
	}
	if err := mode.Validate(); err != nil {
		return core.Job{}, err
	}
	if mode != core.OutputVideo {
		if _, err := core.AudioBitrate(audio.Format, audio.Quality); err != nil {
			return core.Job{}, err
		}
	}
	now := time.Now()
	id := core.JobID(fmt.Sprintf("job-%d-%d", now.UnixMilli(), s.seq.Add(1)))
	j := &core.Job{ID: id, Kind: core.JobDownload, Source: src, Media: &media, Mode: mode, Audio: audio, Video: video, Destination: dest, Status: core.JobPending, Progress: core.JobProgress{Phase: core.PhaseQueued}, CreatedAt: now, UpdatedAt: now}
	result := clone(j)
	s.put(j)
	s.queue <- id
	return result, nil
}
func (s *Service) StartConversions(paths []string, audio core.AudioOptions, dest string) ([]core.Job, error) {
	if _, err := core.AudioBitrate(audio.Format, audio.Quality); err != nil {
		return nil, err
	}
	out := make([]core.Job, 0, len(paths))
	for _, p := range paths {
		info, err := os.Stat(p)
		if err != nil || info.IsDir() {
			continue
		}
		now := time.Now()
		id := core.JobID(fmt.Sprintf("job-%d-%d", now.UnixMilli(), s.seq.Add(1)))
		j := &core.Job{ID: id, Kind: core.JobLocalConversion, InputPath: p, Mode: core.OutputAudio, Audio: audio, Destination: dest, Status: core.JobPending, Progress: core.JobProgress{Phase: core.PhaseQueued}, CreatedAt: now, UpdatedAt: now}
		result := clone(j)
		s.put(j)
		s.queue <- id
		out = append(out, result)
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("no valid input files")
	}
	return out, nil
}
func (s *Service) ListJobs(offset, limit int) []core.Job {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v := make([]core.Job, 0, len(s.jobs))
	for _, j := range s.jobs {
		v = append(v, clone(j))
	}
	sort.Slice(v, func(i, j int) bool { return v[i].CreatedAt.After(v[j].CreatedAt) })
	if offset > len(v) {
		return nil
	}
	v = v[offset:]
	if limit > 0 && len(v) > limit {
		v = v[:limit]
	}
	return v
}
func (s *Service) Cancel(id core.JobID) (core.Job, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	j, ok := s.jobs[id]
	if !ok {
		return core.Job{}, os.ErrNotExist
	}
	if c := s.cancels[id]; c != nil {
		c()
	}
	if j.Status != core.JobCompleted && j.Status != core.JobFailed && j.Status != core.JobCancelled {
		_ = j.Transition(core.JobCancelled, time.Now())
		s.changedLocked(j)
	}
	return clone(j), nil
}
func (s *Service) Retry(id core.JobID) (core.Job, error) {
	s.mu.RLock()
	old, ok := s.jobs[id]
	if !ok {
		s.mu.RUnlock()
		return core.Job{}, os.ErrNotExist
	}
	c := clone(old)
	s.mu.RUnlock()
	var j core.Job
	var err error
	if c.Kind == core.JobDownload && c.Media != nil {
		j, err = s.StartDownload(c.Source, *c.Media, c.Mode, c.Video, c.Audio, c.Destination)
	} else {
		var v []core.Job
		v, err = s.StartConversions([]string{c.InputPath}, c.Audio, c.Destination)
		if len(v) > 0 {
			j = v[0]
		}
	}
	if err == nil {
		s.mu.Lock()
		s.jobs[j.ID].RetryOf = id
		j = clone(s.jobs[j.ID])
		s.mu.Unlock()
	}
	return j, err
}
func (s *Service) worker() {
	for id := range s.queue {
		s.run(id)
	}
}
func (s *Service) run(id core.JobID) {
	s.mu.Lock()
	j := s.jobs[id]
	if j == nil || j.Status == core.JobCancelled {
		s.mu.Unlock()
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	s.cancels[id] = cancel
	if j.Kind == core.JobDownload {
		_ = j.Transition(core.JobDownloading, time.Now())
		j.Progress.Phase = core.PhaseDownloading
	} else {
		_ = j.Transition(core.JobProcessing, time.Now())
		j.Progress.Phase = core.PhaseTranscoding
	}
	s.changedLocked(j)
	s.mu.Unlock()
	var err error
	if j.Kind == core.JobDownload {
		modes := []core.OutputMode{j.Mode}
		if j.Mode == core.OutputBoth {
			modes = []core.OutputMode{core.OutputVideo, core.OutputAudio}
		}
		for _, mode := range modes {
			r := core.DownloadRequest{JobID: id, Source: j.Source, Media: *j.Media, Destination: j.Destination, Mode: mode, Audio: j.Audio, Video: j.Video}
			err = s.downloader.Download(ctx, r, func(e core.Event) { s.event(id, mode, e) })
			if err != nil {
				break
			}
		}
	} else {
		var path string
		path, err = s.converter.Convert(ctx, j.InputPath, j.Destination, j.Audio, func(p core.JobProgress) { s.progress(id, p) })
		if err == nil {
			s.mu.Lock()
			j = s.jobs[id]
			j.Artifacts = append(j.Artifacts, core.Artifact{Kind: core.OutputAudio, Path: path, Available: true})
			s.mu.Unlock()
		}
	}
	s.finish(id, err)
	cancel()
}
func (s *Service) event(id core.JobID, mode core.OutputMode, e core.Event) {
	s.mu.Lock()
	defer s.mu.Unlock()
	j := s.jobs[id]
	switch v := e.(type) {
	case core.DownloadProgress:
		p := float64(0)
		if v.TotalBytes > 0 {
			p = float64(v.DownloadedBytes) / float64(v.TotalBytes)
			j.Progress.Fraction = &p
		}
		j.Progress.CompletedBytes = v.DownloadedBytes
		j.Progress.TotalBytes = v.TotalBytes
		j.Progress.SpeedBytesPerSecond = v.SpeedBytesPerSecond
	case core.DownloadCompleted:
		j.Artifacts = append(j.Artifacts, core.Artifact{Kind: mode, Path: v.Path, Available: true})
	}
	j.UpdatedAt = time.Now()
	s.changedLocked(j)
}
func (s *Service) progress(id core.JobID, p core.JobProgress) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if j := s.jobs[id]; j != nil {
		j.Progress = p
		j.UpdatedAt = time.Now()
		s.changedLocked(j)
	}
}
func (s *Service) finish(id core.JobID, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	j := s.jobs[id]
	delete(s.cancels, id)
	if j.Status == core.JobCancelled {
		return
	}
	if err != nil {
		if errors.Is(err, context.Canceled) {
			_ = j.Transition(core.JobCancelled, time.Now())
		} else {
			_ = j.Transition(core.JobFailed, time.Now())
			j.Failure = &core.JobError{Code: "operation_failed", Message: "Não foi possível concluir a operação.", Details: err.Error()}
		}
	} else {
		j.Progress.Phase = core.PhaseFinalizing
		p := 1.0
		j.Progress.Fraction = &p
		_ = j.Transition(core.JobCompleted, time.Now())
	}
	s.changedLocked(j)
}
func (s *Service) put(j *core.Job) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.jobs[j.ID] = j
	s.changedLocked(j)
}
func (s *Service) changedLocked(j *core.Job) {
	_ = s.persistLocked()
	if s.notify != nil {
		s.notify(clone(j))
	}
}
func clone(j *core.Job) core.Job {
	b, _ := json.Marshal(j)
	var c core.Job
	_ = json.Unmarshal(b, &c)
	return c
}
func (s *Service) persistLocked() error {
	if s.history == "" {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(s.history), 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(s.jobs, "", "  ")
	if err != nil {
		return err
	}
	tmp := s.history + ".tmp"
	f, err := os.OpenFile(tmp, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	if _, err = f.Write(data); err == nil {
		err = f.Sync()
	}
	if e := f.Close(); err == nil {
		err = e
	}
	if err != nil {
		return err
	}
	return os.Rename(tmp, s.history)
}
func (s *Service) load() {
	if s.history == "" {
		return
	}
	b, err := os.ReadFile(s.history)
	if err != nil {
		return
	}
	_ = json.Unmarshal(b, &s.jobs)
	now := time.Now()
	for _, j := range s.jobs {
		if j.Status == core.JobDownloading || j.Status == core.JobProcessing || j.Status == core.JobAnalyzing {
			j.Status = core.JobFailed
			j.Failure = &core.JobError{Code: "interrupted", Message: "Operação interrompida"}
			j.UpdatedAt = now
			t := now
			j.FinishedAt = &t
		}
		for i := range j.Artifacts {
			_, err := os.Stat(j.Artifacts[i].Path)
			j.Artifacts[i].Available = err == nil
		}
	}
	s.mu.Lock()
	_ = s.persistLocked()
	s.mu.Unlock()
}
