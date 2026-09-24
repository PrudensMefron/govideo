package core

import "time"

type Event interface {
	event()
}

type DownloadStarted struct {
	JobID JobID
}

func (DownloadStarted) event() {}

type DownloadProgress struct {
	JobID JobID

	DownloadedBytes int64
	TotalBytes      int64

	SpeedBytesPerSecond float64

	ETA time.Duration
}

func (DownloadProgress) event() {}

type DownloadProcessing struct {
	JobID JobID
	Name  string
}

func (DownloadProcessing) event() {}

type DownloadCompleted struct {
	JobID JobID
	Path  string
}

func (DownloadCompleted) event() {}

type DownloadFailed struct {
	JobID JobID
	Err   error
}

func (DownloadFailed) event() {}
