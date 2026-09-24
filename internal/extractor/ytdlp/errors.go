package ytdlp

import (
	"fmt"
	"strings"
)

type CommandError struct {
	Err    error
	Stderr string
}

func (e *CommandError) Error() string {
	message := strings.TrimSpace(e.Stderr)

	if message == "" {
		return fmt.Sprintf(
			"yt-dlp command failed: %v",
			e.Err,
		)
	}

	return fmt.Sprintf(
		"yt-dlp command failed: %s",
		message,
	)
}

func (e *CommandError) Unwrap() error {
	return e.Err
}
