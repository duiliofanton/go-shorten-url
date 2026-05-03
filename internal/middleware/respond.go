package middleware

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/duiliofanton/go-shorten-url/internal/models"
)

func writeJSONError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(models.ErrorResponse{
		Error:   http.StatusText(status),
		Message: message,
	}); err != nil {
		slog.Error("failed to encode error", "error", err)
	}
}
