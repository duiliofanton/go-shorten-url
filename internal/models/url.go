package models

import (
	"time"
)

type URL struct {
	ID        string    `json:"id"`
	Original  string    `json:"original"`
	ShortCode string    `json:"short_code"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreateURLRequest struct {
	Original string `json:"original" validate:"required,url"`
}

type UpdateURLRequest struct {
	Original string `json:"original" validate:"required,url"`
}

type URLResponse struct {
	ID        string    `json:"id"`
	Original  string    `json:"original"`
	ShortCode string    `json:"short_code"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ListURLsResponse struct {
	URLs       []URLResponse `json:"urls"`
	Total      int          `json:"total"`
	Page       int          `json:"page"`
	PerPage    int          `json:"per_page"`
	TotalPages int          `json:"total_pages"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

type HealthResponse struct {
	Status string `json:"status"`
}