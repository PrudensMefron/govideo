package core

import (
	"fmt"
	"net/url"
	"strings"
)

type Source struct {
	URL     string            `json:"url"`
	PageURL string            `json:"pageURL,omitempty"`
	Headers map[string]string `json:"headers,omitempty"`
}

func (s Source) IsBlob() bool {
	return strings.HasPrefix(strings.ToLower(s.URL), "blob:")
}

func (s Source) Validate() error {
	if strings.TrimSpace(s.URL) == "" {
		return ErrEmptyURL
	}

	if s.IsBlob() {
		if strings.TrimSpace(s.PageURL) == "" {
			return ErrPageURLRequired
		}

		return validateHttpURL(s.PageURL)
	}

	return validateHttpURL(s.URL)
}

func validateHttpURL(raw string) error {
	parsed, err := url.ParseRequestURI(strings.TrimSpace(raw))
	if err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidURL, err)
	}

	switch strings.ToLower(parsed.Scheme) {
	case "http", "https":
		if parsed.Hostname() == "" || parsed.User != nil {
			return fmt.Errorf("%w: URL must contain a host and no credentials", ErrInvalidURL)
		}
		return nil
	default:
		return fmt.Errorf(
			"%w: unsupported scheme %q",
			ErrInvalidURL,
			parsed.Scheme,
		)
	}
}
