package ytdlp

type rawInfo struct {
	Type string `json:"_type"`

	ID         string `json:"id"`
	Title      string `json:"title"`
	FullTitle  string `json:"fulltitle"`
	WebpageURL string `json:"webpage_url"`
	Thumbnail  string `json:"thumbnail"`

	FormatID string `json:"format_id"`
	Ext      string `json:"ext"`

	VCodec string `json:"vcodec"`
	ACodec string `json:"acodec"`

	Duration float64 `json:"duration"`
	TBR      float64 `json:"tbr"`

	Filesize       int64 `json:"filesize"`
	FilesizeApprox int64 `json:"filesize_approx"`

	Width  int `json:"width"`
	Height int `json:"height"`

	HasDRM bool `json:"has_drm"`

	Formats          []rawFormat `json:"formats"`
	RequestedFormats []rawFormat `json:"requested_formats"`

	Entries []rawInfo `json:"entries"`
}

type rawFormat struct {
	ID  string `json:"format_id"`
	Ext string `json:"ext"`

	VCodec string `json:"vcodec"`
	ACodec string `json:"acodec"`

	Filesize       int64 `json:"filesize"`
	FilesizeApprox int64 `json:"filesize_approx"`

	Width  int `json:"width"`
	Height int `json:"height"`

	TBR float64 `json:"tbr"`

	HasDRM bool `json:"has_drm"`
}
