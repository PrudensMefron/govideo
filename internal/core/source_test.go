package core

import (
	"errors"
	"testing"
)

func TestSourceValidateHttp(t *testing.T) {
	source := Source{
		URL: "https://example.com/video",
	}

	if err := source.Validate(); err != nil {
		t.Fatalf("expected valid source, got %v", err)
	}
}

func TestSourceValidateInvalidScheme(t *testing.T) {
	source := Source{
		URL: "ftp://example.com/video",
	}

	err := source.Validate()

	if !errors.Is(err, ErrInvalidURL) {
		t.Fatalf(
			"expected ErrInvalidURL, got %v",
			err,
		)
	}
}

func TestBlobRequiresPageURL(t *testing.T) {
	source := Source{
		URL: "blob:https//example.com/123",
	}

	err := source.Validate()

	if !errors.Is(err, ErrPageURLRequired) {
		t.Fatalf(
			"expected ErrPageURLRequired, got %v",
			err,
		)
	}
}

func TestBlobWithPageURL(t *testing.T) {
	source := Source{
		URL:     "blob:https//example.com/123",
		PageURL: "https://example.com/watch/123",
	}

	if err := source.Validate(); err != nil {
		t.Fatalf("expected valid blob source, got %v", err)
	}
}

func TestSourceRejectsMissingHostAndCredentials(t *testing.T) {
	for _, raw := range []string{"https:///watch", "https://user:pass@example.com/video"} {
		if err := (Source{URL: raw}).Validate(); !errors.Is(err, ErrInvalidURL) {
			t.Errorf("%q: got %v", raw, err)
		}
	}
}
