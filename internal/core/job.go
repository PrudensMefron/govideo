package core

import (
	"fmt"
	"time"
)

type JobID string
type JobStatus string

const (
	JobPending     JobStatus = "pending"
	JobAnalyzing   JobStatus = "analyzing"
	JobReady       JobStatus = "ready"
	JobDownloading JobStatus = "downloading"
	JobProcessing  JobStatus = "processing"
	JobCompleted   JobStatus = "completed"
	JobFailed      JobStatus = "failed"
	JobCancelled   JobStatus = "cancelled"
)

type JobPhase string

const (
	PhasePreparing   JobPhase = "preparing_dependencies"
	PhaseAnalyzing   JobPhase = "analyzing"
	PhaseQueued      JobPhase = "queued"
	PhaseDownloading JobPhase = "downloading"
	PhaseMerging     JobPhase = "merging"
	PhaseExtracting  JobPhase = "extracting_audio"
	PhaseTranscoding JobPhase = "transcoding"
	PhaseFinalizing  JobPhase = "finalizing"
)

type JobProgress struct {
	Phase               JobPhase `json:"phase"`
	Fraction            *float64 `json:"fraction,omitempty"`
	CompletedBytes      int64    `json:"completedBytes,omitempty"`
	TotalBytes          int64    `json:"totalBytes,omitempty"`
	SpeedBytesPerSecond float64  `json:"speedBytesPerSecond,omitempty"`
	ETASeconds          int64    `json:"etaSeconds,omitempty"`
}
type Artifact struct {
	Kind      OutputMode `json:"kind"`
	Path      string     `json:"path"`
	Available bool       `json:"available"`
}
type JobError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}
type Job struct {
	ID          JobID        `json:"id"`
	Kind        JobKind      `json:"kind"`
	Source      Source       `json:"source"`
	InputPath   string       `json:"inputPath,omitempty"`
	Media       *Media       `json:"media,omitempty"`
	Mode        OutputMode   `json:"mode"`
	Audio       AudioOptions `json:"audio"`
	Video       VideoQuality `json:"video"`
	Destination string       `json:"destination"`
	Status      JobStatus    `json:"status"`
	Progress    JobProgress  `json:"progress"`
	Artifacts   []Artifact   `json:"artifacts,omitempty"`
	Failure     *JobError    `json:"failure,omitempty"`
	RetryOf     JobID        `json:"retryOf,omitempty"`
	CreatedAt   time.Time    `json:"createdAt"`
	UpdatedAt   time.Time    `json:"updatedAt"`
	StartedAt   *time.Time   `json:"startedAt,omitempty"`
	FinishedAt  *time.Time   `json:"finishedAt,omitempty"`
}

var transitions = map[JobStatus]map[JobStatus]bool{
	JobPending:   {JobAnalyzing: true, JobReady: true, JobDownloading: true, JobProcessing: true, JobCancelled: true, JobFailed: true},
	JobAnalyzing: {JobReady: true, JobFailed: true, JobCancelled: true}, JobReady: {JobDownloading: true, JobProcessing: true, JobCancelled: true, JobFailed: true},
	JobDownloading: {JobProcessing: true, JobCompleted: true, JobFailed: true, JobCancelled: true}, JobProcessing: {JobCompleted: true, JobFailed: true, JobCancelled: true},
}

func (j *Job) Transition(next JobStatus, now time.Time) error {
	if j.Status == next {
		return nil
	}
	if !transitions[j.Status][next] {
		return fmt.Errorf("%w: %s -> %s", ErrInvalidTransition, j.Status, next)
	}
	j.Status = next
	j.UpdatedAt = now
	if j.StartedAt == nil && (next == JobAnalyzing || next == JobDownloading || next == JobProcessing) {
		t := now
		j.StartedAt = &t
	}
	if next == JobCompleted || next == JobFailed || next == JobCancelled {
		t := now
		j.FinishedAt = &t
	}
	return nil
}

type DownloadRequest struct {
	JobID       JobID
	Source      Source
	Media       Media
	Destination string
	Mode        OutputMode
	Audio       AudioOptions
	Video       VideoQuality
}
