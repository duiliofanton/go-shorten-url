package validator

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateURL(t *testing.T) {
	tests := []struct {
		name    string
		url    string
		errNil bool
	}{
		{"valid https", "https://google.com", true},
		{"valid http", "http://example.com", true},
		{"valid with path", "https://google.com/search?q=test", true},
		{"empty", "", false},
		{"no scheme", "google.com", false},
		{"invalid scheme", "ftp://example.com", false},
		{"just scheme", "http://", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateURL(tt.url)
			if tt.errNil {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}