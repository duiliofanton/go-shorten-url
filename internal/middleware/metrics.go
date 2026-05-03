package middleware

import (
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

// uniqueIPWindow caps the memory footprint of unique-IP tracking by clearing
// the set periodically. The exposed metric is "unique IPs in the last window".
const uniqueIPWindow = time.Hour

type Metrics struct {
	requestsTotal uint64
	statusBuckets [6]uint64 // index 1..5 for 1xx..5xx; index 0 for unknown

	uniqueIPsMu sync.Mutex
	uniqueIPs   map[string]struct{}

	serverStart time.Time
	done        chan struct{}
}

func NewMetrics() *Metrics {
	m := &Metrics{
		uniqueIPs:   make(map[string]struct{}),
		serverStart: time.Now(),
		done:        make(chan struct{}),
	}
	go m.rotate()
	return m
}

// Stop terminates the background rotation goroutine.
func (m *Metrics) Stop() {
	close(m.done)
}

func (m *Metrics) rotate() {
	ticker := time.NewTicker(uniqueIPWindow)
	defer ticker.Stop()
	for {
		select {
		case <-m.done:
			return
		case <-ticker.C:
			m.uniqueIPsMu.Lock()
			m.uniqueIPs = make(map[string]struct{})
			m.uniqueIPsMu.Unlock()
		}
	}
}

func (m *Metrics) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(rw, r)

		atomic.AddUint64(&m.requestsTotal, 1)
		atomic.AddUint64(&m.statusBuckets[bucket(rw.statusCode)], 1)

		ip := clientIP(r)
		m.uniqueIPsMu.Lock()
		m.uniqueIPs[ip] = struct{}{}
		m.uniqueIPsMu.Unlock()
	})
}

func bucket(status int) int {
	b := status / 100
	if b < 1 || b > 5 {
		return 0
	}
	return b
}

func (m *Metrics) Handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")

		total := atomic.LoadUint64(&m.requestsTotal)
		s2xx := atomic.LoadUint64(&m.statusBuckets[2])
		s3xx := atomic.LoadUint64(&m.statusBuckets[3])
		s4xx := atomic.LoadUint64(&m.statusBuckets[4])
		s5xx := atomic.LoadUint64(&m.statusBuckets[5])

		m.uniqueIPsMu.Lock()
		uniqueIPs := len(m.uniqueIPs)
		m.uniqueIPsMu.Unlock()

		uptime := int(time.Since(m.serverStart).Seconds())

		_, err := fmt.Fprintf(w,
			"# HELP requests_total Total HTTP requests\n"+
				"# TYPE requests_total counter\n"+
				"requests_total %d\n\n"+
				"# HELP requests_by_status HTTP requests by status code class\n"+
				"# TYPE requests_by_status counter\n"+
				"requests_by_status{code=\"2xx\"} %d\n"+
				"requests_by_status{code=\"3xx\"} %d\n"+
				"requests_by_status{code=\"4xx\"} %d\n"+
				"requests_by_status{code=\"5xx\"} %d\n\n"+
				"# HELP unique_ips Unique IP addresses in the last hour\n"+
				"# TYPE unique_ips gauge\n"+
				"unique_ips %d\n\n"+
				"# HELP server_uptime_seconds Server uptime in seconds\n"+
				"# TYPE server_uptime_seconds gauge\n"+
				"server_uptime_seconds %d\n",
			total, s2xx, s3xx, s4xx, s5xx, uniqueIPs, uptime,
		)
		if err != nil {
			slog.Error("failed to write metrics", "error", err)
		}
	}
}
