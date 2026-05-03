package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/duiliofanton/go-shorten-url/internal/models"
	"github.com/duiliofanton/go-shorten-url/internal/service"
	"github.com/stretchr/testify/assert"
)

type mockRedirectService struct {
	resp *models.URLResponse
	err  error
}

func (m *mockRedirectService) Create(ctx context.Context, req *models.CreateURLRequest) (*models.URLResponse, error) {
	return m.resp, m.err
}

func (m *mockRedirectService) GetByID(ctx context.Context, id string) (*models.URLResponse, error) {
	return m.resp, m.err
}

func (m *mockRedirectService) GetByShortCode(ctx context.Context, shortCode string) (*models.URLResponse, error) {
	if m.err != nil {
		return nil, m.err
	}
	if m.resp != nil {
		return m.resp, nil
	}
	return nil, service.ErrURLNotFound
}

func (m *mockRedirectService) Update(ctx context.Context, id string, req *models.UpdateURLRequest) (*models.URLResponse, error) {
	return m.resp, m.err
}

func (m *mockRedirectService) Delete(ctx context.Context, id string) error {
	return m.err
}

func (m *mockRedirectService) List(ctx context.Context, page, perPage int) (*models.ListURLsResponse, error) {
	return &models.ListURLsResponse{}, m.err
}

func setupRedirectTestRouter(svc *mockRedirectService) *http.ServeMux {
	handler := NewRedirectHandler(svc)
	mux := http.NewServeMux()
	mux.HandleFunc("GET /{shortCode}", handler.Redirect)
	return mux
}

func TestRedirect_Success(t *testing.T) {
	svc := &mockRedirectService{
		resp: &models.URLResponse{
			ID:        "test-id",
			Original:  "https://google.com",
			ShortCode: "abc123",
		},
	}
	mux := setupRedirectTestRouter(svc)

	req := httptest.NewRequest(http.MethodGet, "/abc123", nil)
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	assert.Equal(t, http.StatusFound, w.Code)
	assert.Equal(t, "https://google.com", w.Header().Get("Location"))
}

func TestRedirect_NotFound(t *testing.T) {
	svc := &mockRedirectService{
		err: service.ErrURLNotFound,
	}
	mux := setupRedirectTestRouter(svc)

	req := httptest.NewRequest(http.MethodGet, "/notfound", nil)
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestRedirect_InternalError(t *testing.T) {
	svc := &mockRedirectService{
		err: errors.New("internal error"),
	}
	mux := setupRedirectTestRouter(svc)

	req := httptest.NewRequest(http.MethodGet, "/abc123", nil)
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestRedirect_EmptyShortCode(t *testing.T) {
	svc := &mockRedirectService{}
	mux := setupRedirectTestRouter(svc)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}
