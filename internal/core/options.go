package core

import "fmt"

type OutputMode string

const (
	OutputVideo OutputMode = "video"
	OutputAudio OutputMode = "audio"
	OutputBoth  OutputMode = "both"
)

type JobKind string

const (
	JobDownload        JobKind = "download"
	JobLocalConversion JobKind = "local_conversion"
)

type AudioFormat string

const (
	AudioMP3  AudioFormat = "mp3"
	AudioM4A  AudioFormat = "m4a"
	AudioOpus AudioFormat = "opus"
)

type AudioQuality string

const (
	AudioEconomy  AudioQuality = "economy"
	AudioStandard AudioQuality = "standard"
	AudioHigh     AudioQuality = "high"
)

type VideoQuality struct {
	MaxHeight int `json:"maxHeight"`
}
type AudioOptions struct {
	Format  AudioFormat  `json:"format"`
	Quality AudioQuality `json:"quality"`
}

func (m OutputMode) Validate() error {
	if m != OutputVideo && m != OutputAudio && m != OutputBoth {
		return fmt.Errorf("%w: output mode %q", ErrInvalidOption, m)
	}
	return nil
}
func (f AudioFormat) Validate() error {
	if f != AudioMP3 && f != AudioM4A && f != AudioOpus {
		return fmt.Errorf("%w: audio format %q", ErrInvalidOption, f)
	}
	return nil
}
func (q AudioQuality) Validate() error {
	if q != AudioEconomy && q != AudioStandard && q != AudioHigh {
		return fmt.Errorf("%w: audio quality %q", ErrInvalidOption, q)
	}
	return nil
}
func AudioBitrate(format AudioFormat, quality AudioQuality) (int, error) {
	if err := format.Validate(); err != nil {
		return 0, err
	}
	if err := quality.Validate(); err != nil {
		return 0, err
	}
	values := map[AudioFormat]map[AudioQuality]int{AudioMP3: {AudioEconomy: 128, AudioStandard: 192, AudioHigh: 320}, AudioM4A: {AudioEconomy: 96, AudioStandard: 128, AudioHigh: 256}, AudioOpus: {AudioEconomy: 64, AudioStandard: 96, AudioHigh: 160}}
	return values[format][quality], nil
}
