package ytdlp

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/PrudensMefron/govideo/internal/core"
	"github.com/PrudensMefron/govideo/internal/process"
)

const (
	defaultBinary = "yt-dlp"
	defaultFormat = "bv*+ba/b"
)

type Adapter struct {
	binary string
	format string
}

func New(binary string) *Adapter {
	if strings.TrimSpace(binary) == "" {
		binary = defaultBinary
	}

	return &Adapter{
		binary: binary,
		format: defaultFormat,
	}
}

func Default() *Adapter {
	return New(defaultBinary)
}

var _ core.Extractor = (*Adapter)(nil)

func (a *Adapter) Probe(
	ctx context.Context,
	source core.Source,
) (core.Media, error) {
	if err := source.Validate(); err != nil {
		return core.Media{}, err
	}

	if source.IsBlob() {
		return core.Media{}, fmt.Errorf(
			"%w: blob URL requires browser discovery",
			core.ErrUnsupportedSource,
		)
	}

	args := a.probeArgs(source)

	cmd := process.CommandContext(
		ctx,
		a.binary,
		args...,
	)

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		if errors.Is(ctx.Err(), context.Canceled) {
			return core.Media{}, ctx.Err()
		}

		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return core.Media{}, ctx.Err()
		}

		return core.Media{}, &CommandError{
			Err:    err,
			Stderr: stderr.String(),
		}
	}

	var info rawInfo

	if err := json.Unmarshal(stdout.Bytes(), &info); err != nil {
		return core.Media{}, fmt.Errorf(
			"decode yt-dlp response: %w",
			err,
		)
	}

	info, err := normalizeInfo(info)
	if err != nil {
		return core.Media{}, err
	}

	return mapMedia(info), nil
}

func (a *Adapter) probeArgs(source core.Source) []string {
	args := []string{
		"--dump-single-json",
		"--skip-download",
		"--no-playlist",
		"--check-formats",
		"--no-warnings",
		"--socket-timeout",
		"20",
		"--format",
		a.format,
	}

	args = appendHeaders(
		args,
		source.Headers,
	)

	args = append(args, source.URL)

	return args
}

func appendHeaders(args []string, headers map[string]string) []string {
	if len(headers) == 0 {
		return args
	}

	keys := make([]string, 0, len(headers))

	for key := range headers {
		keys = append(keys, key)
	}

	sort.Strings(keys)

	for _, key := range keys {
		value := headers[key]

		if strings.TrimSpace(key) == "" {
			continue
		}

		args = append(
			args,
			"--add-headers",
			fmt.Sprintf("%s:%s", key, value),
		)
	}

	return args
}

func normalizeInfo(info rawInfo) (rawInfo, error) {
	if info.Type != "playlist" {
		return info, nil
	}

	for _, entry := range info.Entries {
		if entry.ID != "" || entry.Title != "" {
			return entry, nil
		}
	}

	return rawInfo{}, core.ErrMediaNotFound
}
