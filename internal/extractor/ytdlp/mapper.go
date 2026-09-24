package ytdlp

import (
	"time"

	"github.com/PrudensMefron/govideo/internal/core"
)

func mapMedia(info rawInfo) core.Media {
	title := info.Title

	if title == "" {
		title = info.FullTitle
	}

	size, estimated := estimateSize(info)

	formats := make([]core.Format, 0, len(info.Formats))

	for _, format := range info.Formats {
		formats = append(formats, mapFormat(format))
	}

	return core.Media{
		ID:           info.ID,
		Title:        title,
		WebpageURL:   info.WebpageURL,
		ThumbnailURL: info.Thumbnail,

		Duration: time.Duration(
			info.Duration * float64(time.Second),
		),

		SizeBytes:     size,
		SizeEstimated: estimated,

		DRM: hasDRM(info),

		Formats: formats,
	}
}

func mapFormat(format rawFormat) core.Format {
	return core.Format{
		ID:        format.ID,
		Extension: format.Ext,

		Width:  format.Width,
		Height: format.Height,

		BitrateKbps: format.TBR,

		VideoCodec: format.VCodec,
		AudioCodec: format.ACodec,

		HasVideo: hasCodec(format.VCodec),
		HasAudio: hasCodec(format.ACodec),
	}
}

func hasCodec(codec string) bool {
	return codec != "" && codec != "none"
}

func hasDRM(info rawInfo) bool {
	if info.HasDRM {
		return true
	}

	for _, format := range selectedFormats(info) {
		if format.HasDRM {
			return true
		}
	}

	return false
}

func selectedFormats(info rawInfo) []rawFormat {
	if len(info.RequestedFormats) > 0 {
		return info.RequestedFormats
	}

	return []rawFormat{
		{
			ID:             info.FormatID,
			Ext:            info.Ext,
			Width:          info.Width,
			Height:         info.Height,
			TBR:            info.TBR,
			VCodec:         info.VCodec,
			ACodec:         info.ACodec,
			Filesize:       info.Filesize,
			FilesizeApprox: info.FilesizeApprox,
			HasDRM:         info.HasDRM,
		},
	}
}

func estimateSize(info rawInfo) (int64, bool) {
	formats := selectedFormats(info)

	var total int64
	var found bool
	var estimated bool

	for _, format := range formats {
		size, isEstimate := formatSize(
			format,
			info.Duration,
		)

		if size <= 0 {
			continue
		}

		total += size
		found = true

		if isEstimate {
			estimated = true
		}
	}

	if !found {
		return 0, true
	}

	return total, estimated
}

func formatSize(format rawFormat, duration float64) (int64, bool) {
	if format.Filesize > 0 {
		return format.Filesize, false
	}

	if format.FilesizeApprox > 0 {
		return format.FilesizeApprox, true
	}

	if format.TBR > 0 && duration > 0 {
		size := duration * format.TBR * 1000 / 8

		return int64(size), true
	}

	return 0, true
}
