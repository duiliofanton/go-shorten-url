package handler

import (
	"errors"
	"net/http"

	"github.com/duiliofanton/go-shorten-url/internal/service"
)

type RedirectHandler struct {
	service service.URLService
}

func NewRedirectHandler(svc service.URLService) *RedirectHandler {
	return &RedirectHandler{service: svc}
}

func (h *RedirectHandler) Redirect(w http.ResponseWriter, r *http.Request) {
	shortCode := r.PathValue("shortCode")
	if shortCode == "" {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	url, err := h.service.GetByShortCode(r.Context(), shortCode)
	if err != nil {
		if errors.Is(err, service.ErrURLNotFound) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, url.Original, http.StatusFound)
}