package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/duiliofanton/go-shorten-url/internal/models"
	"github.com/duiliofanton/go-shorten-url/internal/service"
)

const maxRequestBody = 16 << 10 // 16 KiB

type URLHandler struct {
	service service.URLService
}

func NewURLHandler(svc service.URLService) *URLHandler {
	return &URLHandler{service: svc}
}

func (h *URLHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req models.CreateURLRequest
	if err := decodeJSON(w, r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	url, err := h.service.Create(r.Context(), &req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidURL):
			WriteError(w, http.StatusBadRequest, "invalid URL format")
		default:
			WriteError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	WriteJSON(w, http.StatusCreated, url)
}

func (h *URLHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		WriteError(w, http.StatusBadRequest, "id is required")
		return
	}

	url, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrURLNotFound) {
			WriteError(w, http.StatusNotFound, "url not found")
			return
		}
		WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	WriteJSON(w, http.StatusOK, url)
}

func (h *URLHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		WriteError(w, http.StatusBadRequest, "id is required")
		return
	}

	var req models.UpdateURLRequest
	if err := decodeJSON(w, r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	url, err := h.service.Update(r.Context(), id, &req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidURL):
			WriteError(w, http.StatusBadRequest, "invalid URL format")
		case errors.Is(err, service.ErrURLNotFound):
			WriteError(w, http.StatusNotFound, "url not found")
		default:
			WriteError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	WriteJSON(w, http.StatusOK, url)
}

func (h *URLHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		WriteError(w, http.StatusBadRequest, "id is required")
		return
	}

	if err := h.service.Delete(r.Context(), id); err != nil {
		if errors.Is(err, service.ErrURLNotFound) {
			WriteError(w, http.StatusNotFound, "url not found")
			return
		}
		WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *URLHandler) List(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	perPage, _ := strconv.Atoi(r.URL.Query().Get("per_page"))

	urls, err := h.service.List(r.Context(), page, perPage)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	WriteJSON(w, http.StatusOK, urls)
}

func decodeJSON(w http.ResponseWriter, r *http.Request, v any) error {
	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBody)
	return json.NewDecoder(r.Body).Decode(v)
}

func WriteJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		slog.Error("failed to encode json", "error", err)
	}
}

func WriteError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(models.ErrorResponse{
		Error:   http.StatusText(status),
		Message: message,
	}); err != nil {
		slog.Error("failed to encode json", "error", err)
	}
}
