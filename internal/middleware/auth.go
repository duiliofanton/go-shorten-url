package middleware

import (
	"crypto/subtle"
	"net/http"
)

type Auth struct {
	keys [][]byte
}

func NewAuth(keys ...string) *Auth {
	bs := make([][]byte, 0, len(keys))
	for _, k := range keys {
		if k != "" {
			bs = append(bs, []byte(k))
		}
	}
	return &Auth{keys: bs}
}

// Middleware authenticates requests via X-API-Key (or `api_key` query param).
// allowAnonymous bypasses authentication when it returns true; pass nil to require
// authentication on every request.
func (a *Auth) Middleware(allowAnonymous func(*http.Request) bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if allowAnonymous != nil && allowAnonymous(r) {
				next.ServeHTTP(w, r)
				return
			}

			key := r.Header.Get("X-API-Key")
			if key == "" {
				key = r.URL.Query().Get("api_key")
			}
			if key == "" {
				writeJSONError(w, http.StatusUnauthorized, "missing API key")
				return
			}
			if !a.isValid(key) {
				writeJSONError(w, http.StatusUnauthorized, "invalid API key")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// isValid evaluates every configured key without short-circuiting,
// to avoid leaking the position of a valid key via timing.
func (a *Auth) isValid(key string) bool {
	provided := []byte(key)
	valid := false
	for _, vk := range a.keys {
		if subtle.ConstantTimeCompare(provided, vk) == 1 {
			valid = true
		}
	}
	return valid
}
