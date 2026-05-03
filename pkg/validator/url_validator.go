package validator

import (
	"net/url"
	"strings"
)

func ValidateURL(rawURL string) error {
	if rawURL == "" {
		return ErrEmptyURL
	}

	parsed, err := url.Parse(rawURL)
	if err != nil {
		return ErrInvalidURL
	}

	if parsed.Scheme == "" || parsed.Host == "" {
		return ErrInvalidURL
	}

	scheme := strings.ToLower(parsed.Scheme)
	if scheme != "http" && scheme != "https" {
		return ErrInvalidScheme
	}

	return nil
}

var (
	ErrEmptyURL      = &ValidationError{msg: "URL cannot be empty"}
	ErrInvalidURL   = &ValidationError{msg: "invalid URL format"}
	ErrInvalidScheme = &ValidationError{msg: "URL must use http or https scheme"}
)

type ValidationError struct {
	msg string
}

func (e *ValidationError) Error() string {
	return e.msg
}