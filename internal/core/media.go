package core

import "time"

type Media struct {
	ID           string `json:"id"`
	Title        string `json:"title"`
	WebpageURL   string `json:"webpageURL"`
	ThumbnailURL string `json:"thumbnailURL"`

	Duration time.Duration `json:"duration"`

	SizeBytes int64 `json:"sizeBytes"`

	SizeEstimated bool `json:"sizeEstimated"`
	DRM           bool `json:"drm"`

	Formats []Format `json:"formats"`
}

type Format struct {
	ID         string `json:"id"`
	Extension  string `json:"extension"`
	VideoCodec string `json:"videoCodec"`
	AudioCodec string `json:"audioCodec"`

	Width  int `json:"width"`
	Height int `json:"height"`

	BitrateKbps float64 `json:"bitrateKbps"`

	HasVideo bool `json:"hasVideo"`
	HasAudio bool `json:"hasAudio"`
}
