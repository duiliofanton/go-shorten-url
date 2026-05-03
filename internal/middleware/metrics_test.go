package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMetrics_Format(t *testing.T) {
	metrics := NewMetrics()

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	w := httptest.NewRecorder()

	metrics.Handler().ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "text/plain; version=0.0.4; charset=utf-8", w.Header().Get("Content-Type"))

	body := w.Body.String()
	assert.True(t, strings.Contains(body, "# HELP requests_total Total HTTP requests"))
	assert.True(t, strings.Contains(body, "# TYPE requests_total counter"))
	assert.True(t, strings.Contains(body, "requests_total"))
	assert.True(t, strings.Contains(body, "# HELP requests_by_status HTTP requests by status code"))
	assert.True(t, strings.Contains(body, "requests_by_status"))
	assert.True(t, strings.Contains(body, "# HELP unique_ips Unique IP addresses"))
	assert.True(t, strings.Contains(body, "# HELP server_uptime_seconds Server uptime in seconds"))
}

func TestMetrics_Incr(t *testing.T) {
	metrics := NewMetrics()

	for i := 0; i < 5; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = "192.168.1.1:1234"
		w := httptest.NewRecorder()
		w.WriteHeader(http.StatusOK)
		metrics.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})).ServeHTTP(w, req)
	}

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	w := httptest.NewRecorder()

	metrics.Handler().ServeHTTP(w, req)

	body := w.Body.String()
	assert.True(t, strings.Contains(body, "requests_total 5"))
	assert.True(t, strings.Contains(body, `requests_by_status{code="2xx"} 5`))
}

func TestMetrics_UniqueIPs(t *testing.T) {
	metrics := NewMetrics()

	ips := []string{"192.168.1.1", "192.168.1.2", "192.168.1.1"}

	for _, ip := range ips {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = ip + ":1234"
		w := httptest.NewRecorder()
		metrics.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})).ServeHTTP(w, req)
	}

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	w := httptest.NewRecorder()

	metrics.Handler().ServeHTTP(w, req)

	body := w.Body.String()
	assert.True(t, strings.Contains(body, "unique_ips 2"))
}

func TestMetrics_Uptime(t *testing.T) {
	metrics := NewMetrics()

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	w := httptest.NewRecorder()

	metrics.Handler().ServeHTTP(w, req)

	body := w.Body.String()
	assert.True(t, strings.Contains(body, "server_uptime_seconds"))
}

func TestMetrics_StatusCodes(t *testing.T) {
	metrics := NewMetrics()

	statuses := []int{200, 201, 400, 401, 404, 500, 502}

	for _, status := range statuses {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = "192.168.1.1:1234"
		w := httptest.NewRecorder()
		metrics.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(status)
		})).ServeHTTP(w, req)
	}

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	w := httptest.NewRecorder()

	metrics.Handler().ServeHTTP(w, req)

	body := w.Body.String()
	assert.Contains(t, body, "requests_by_status")
}
