package middleware

import (
	"net"
	"net/http"
)

// clientIP returns the IP address of the client, stripping any port.
// X-Forwarded-For is intentionally not consulted: it is trivially spoofable
// unless the server is behind a trusted proxy. Configure your reverse proxy
// to rewrite RemoteAddr (or extend this function with a trusted-proxy list).
func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
