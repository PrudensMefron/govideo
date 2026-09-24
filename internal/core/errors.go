package core

import "errors"

var (
	ErrEmptyURL          = errors.New("empty URL")
	ErrInvalidURL        = errors.New("invalid URL")
	ErrPageURLRequired   = errors.New("page URL required")
	ErrDRMProtected      = errors.New("media is DRM protected")
	ErrMediaNotFound     = errors.New("media not found")
	ErrUnsupportedSource = errors.New("unsupported source")
	ErrInvalidTransition = errors.New("invalid job transition")
	ErrInvalidOption     = errors.New("invalid option")
)
