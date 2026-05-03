package handler

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/duiliofanton/go-shorten-url/internal/models"
	"github.com/duiliofanton/go-shorten-url/internal/service"
	"github.com/stretchr/testify/assert"
)

type mockURLService struct {
	createErr   error
	getByIDErr  error
	updateErr   error
	deleteErr  error
	listResp   *models.ListURLsResponse
	createResp *models.URLResponse
}

func (m *mockURLService) Create(ctx context.Context, req *models.CreateURLRequest) (*models.URLResponse, error) {
	if m.createErr != nil {
		return nil, m.createErr
	}
	return m.createResp, nil
}

func (m *mockURLService) GetByID(ctx context.Context, id string) (*models.URLResponse, error) {
	if m.getByIDErr != nil {
		return nil, m.getByIDErr
	}
	if m.createResp != nil {
		return m.createResp, nil
	}
	return nil, errors.New("not found")
}

func (m *mockURLService) GetByShortCode(ctx context.Context, shortCode string) (*models.URLResponse, error) {
	if m.createResp != nil {
		return m.createResp, nil
	}
	return nil, errors.New("not found")
}

func (m *mockURLService) Update(ctx context.Context, id string, req *models.UpdateURLRequest) (*models.URLResponse, error) {
	if m.updateErr != nil {
		return nil, m.updateErr
	}
	if m.createResp != nil {
		return m.createResp, nil
	}
	return nil, errors.New("not found")
}

func (m *mockURLService) Delete(ctx context.Context, id string) error {
	return m.deleteErr
}

func (m *mockURLService) List(ctx context.Context, page, perPage int) (*models.ListURLsResponse, error) {
	if m.listResp != nil {
		return m.listResp, nil
	}
	return &models.ListURLsResponse{}, nil
}

func setupTestRouter(svc *mockURLService) *http.ServeMux {
	handler := NewURLHandler(svc)
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/urls", handler.Create)
	mux.HandleFunc("GET /api/urls/{id}", handler.GetByID)
	mux.HandleFunc("PUT /api/urls/{id}", handler.Update)
	mux.HandleFunc("DELETE /api/urls/{id}", handler.Delete)
	mux.HandleFunc("GET /api/urls", handler.List)
	return mux
}

func TestCreate_Success(t *testing.T) {
	svc := &mockURLService{
		createResp: &models.URLResponse{
			ID:        "test-id",
			Original:  "https://google.com",
			ShortCode: "abc123",
		},
	}
	mux := setupTestRouter(svc)

	body := []byte(`{"original": "https://google.com"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/urls", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestCreate_InvalidBody(t *testing.T) {
	svc := &mockURLService{}
	mux := setupTestRouter(svc)

	body := []byte(`invalid json`)
	req := httptest.NewRequest(http.MethodPost, "/api/urls", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreate_InvalidURL(t *testing.T) {
	svc := &mockURLService{
		createErr: service.ErrInvalidURL,
	}
	mux := setupTestRouter(svc)

	body := []byte(`{"original": "not-a-url"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/urls", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetByID_Success(t *testing.T) {
	svc := &mockURLService{
		createResp: &models.URLResponse{
			ID:        "test-id",
			Original:  "https://google.com",
			ShortCode: "abc123",
		},
	}
	mux := setupTestRouter(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/urls/test-id", nil)
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGetByID_NotFound(t *testing.T) {
	svc := &mockURLService{
		getByIDErr: service.ErrURLNotFound,
	}
	mux := setupTestRouter(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/urls/not-found", nil)
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestUpdate_Success(t *testing.T) {
	svc := &mockURLService{
		createResp: &models.URLResponse{
			ID:        "test-id",
			Original:  "https://updated.com",
			ShortCode: "abc123",
		},
	}
	mux := setupTestRouter(svc)

	body := []byte(`{"original": "https://updated.com"}`)
	req := httptest.NewRequest(http.MethodPut, "/api/urls/test-id", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestUpdate_NotFound(t *testing.T) {
	svc := &mockURLService{
		updateErr: service.ErrURLNotFound,
	}
	mux := setupTestRouter(svc)

	body := []byte(`{"original": "https://updated.com"}`)
	req := httptest.NewRequest(http.MethodPut, "/api/urls/not-found", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestUpdate_InvalidURL(t *testing.T) {
	svc := &mockURLService{
		updateErr: service.ErrInvalidURL,
	}
	mux := setupTestRouter(svc)

	body := []byte(`{"original": "not-a-url"}`)
	req := httptest.NewRequest(http.MethodPut, "/api/urls/test-id", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDelete_Success(t *testing.T) {
	svc := &mockURLService{}
	mux := setupTestRouter(svc)

	req := httptest.NewRequest(http.MethodDelete, "/api/urls/test-id", nil)
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestDelete_NotFound(t *testing.T) {
	svc := &mockURLService{
		deleteErr: service.ErrURLNotFound,
	}
	mux := setupTestRouter(svc)

	req := httptest.NewRequest(http.MethodDelete, "/api/urls/not-found", nil)
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestList_Success(t *testing.T) {
	svc := &mockURLService{
		listResp: &models.ListURLsResponse{
			URLs:       []models.URLResponse{{ID: "1", Original: "https://google.com"}},
			Total:      1,
			Page:       1,
			PerPage:    10,
			TotalPages: 1,
		},
	}
	mux := setupTestRouter(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/urls", nil)
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHealthCheck(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		WriteJSON(w, http.StatusOK, models.HealthResponse{Status: "ok"})
	})

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}