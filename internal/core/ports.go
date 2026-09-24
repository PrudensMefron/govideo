package core

import "context"

type Extractor interface {
	Probe(
		ctx context.Context,
		source Source,
	) (Media, error)
}

type EventHandler func(Event)

type Downloader interface {
	Download(
		ctx context.Context,
		request DownloadRequest,
		emit EventHandler,
	) error
}
