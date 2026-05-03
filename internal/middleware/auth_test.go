package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func okHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}

func TestAuth_ValidKey(t *testing.T) {
	auth := NewAuth("valid-key", "another-key")
	handler := auth.Middleware(nil)(okHandler())

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-API-Key", "valid-key")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAuth_SecondaryKey(t *testing.T) {
	auth := NewAuth("first-key", "second-key")
	handler := auth.Middleware(nil)(okHandler())

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-API-Key", "second-key")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAuth_InvalidKey(t *testing.T) {
	auth := NewAuth("valid-key")
	handler := auth.Middleware(nil)(okHandler())

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-API-Key", "invalid-key")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
}

func TestAuth_MissingKey(t *testing.T) {
	auth := NewAuth("valid-key")
	handler := auth.Middleware(nil)(okHandler())

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuth_QueryParam(t *testing.T) {
	auth := NewAuth("valid-key")
	handler := auth.Middleware(nil)(okHandler())

	req := httptest.NewRequest(http.MethodGet, "/test?api_key=valid-key", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAuth_AllowAnonymous(t *testing.T) {
	auth := NewAuth("valid-key")
	allow := func(r *http.Request) bool {
		return r.URL.Path == "/health"
	}
	handler := auth.Middleware(allow)(okHandler())

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAuth_AllowAnonymousFalseStillRequiresKey(t *testing.T) {
	auth := NewAuth("valid-key")
	allow := func(r *http.Request) bool {
		return r.URL.Path == "/health"
	}
	handler := auth.Middleware(allow)(okHandler())

	req := httptest.NewRequest(http.MethodGet, "/api/urls", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
