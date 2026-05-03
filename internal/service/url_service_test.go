package service

import (
	"context"
	"testing"

	"github.com/duiliofanton/go-shorten-url/internal/models"
	"github.com/stretchr/testify/assert"
)

type mockRepository struct {
	urls  map[string]*models.URL
	find map[string]*models.URL
}

func newMockRepository() *mockRepository {
	return &mockRepository{
		urls: make(map[string]*models.URL),
		find: make(map[string]*models.URL),
	}
}

func (m *mockRepository) Create(ctx context.Context, url *models.URL) error {
	m.urls[url.ID] = url
	m.find[url.ShortCode] = url
	return nil
}

func (m *mockRepository) GetByID(ctx context.Context, id string) (*models.URL, error) {
	if url, ok := m.urls[id]; ok {
		return url, nil
	}
	return nil, ErrURLNotFound
}

func (m *mockRepository) GetByShortCode(ctx context.Context, shortCode string) (*models.URL, error) {
	if url, ok := m.find[shortCode]; ok {
		return url, nil
	}
	return nil, ErrURLNotFound
}

func (m *mockRepository) Update(ctx context.Context, url *models.URL) error {
	if _, ok := m.urls[url.ID]; !ok {
		return ErrURLNotFound
	}
	m.urls[url.ID] = url
	m.find[url.ShortCode] = url
	return nil
}

func (m *mockRepository) Delete(ctx context.Context, id string) error {
	if _, ok := m.urls[id]; !ok {
		return ErrURLNotFound
	}
	url := m.urls[id]
	delete(m.urls, id)
	delete(m.find, url.ShortCode)
	return nil
}

func (m *mockRepository) List(ctx context.Context, page, perPage int) ([]models.URL, int, error) {
	var all []models.URL
	for _, url := range m.urls {
		all = append(all, *url)
	}

	total := len(all)
	offset := (page - 1) * perPage
	if offset >= total {
		return []models.URL{}, total, nil
	}

	end := offset + perPage
	if end > total {
		end = total
	}

	return all[offset:end], total, nil
}

func TestCreate(t *testing.T) {
	repo := newMockRepository()
	svc := NewURLService(repo)

	req := &models.CreateURLRequest{Original: "https://google.com"}
	resp, err := svc.Create(context.Background(), req)

	assert.NoError(t, err)
	assert.NotEmpty(t, resp.ID)
	assert.Equal(t, "https://google.com", resp.Original)
	assert.NotEmpty(t, resp.ShortCode)
}

func TestCreateInvalidURL(t *testing.T) {
	repo := newMockRepository()
	svc := NewURLService(repo)

	req := &models.CreateURLRequest{Original: "not-a-url"}
	resp, err := svc.Create(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, ErrInvalidURL, err)
}

func TestGetByID(t *testing.T) {
	repo := newMockRepository()
	svc := NewURLService(repo)

	createReq := &models.CreateURLRequest{Original: "https://google.com"}
	created, _ := svc.Create(context.Background(), createReq)

	resp, err := svc.GetByID(context.Background(), created.ID)

	assert.NoError(t, err)
	assert.Equal(t, created.ID, resp.ID)
}

func TestGetByIDNotFound(t *testing.T) {
	repo := newMockRepository()
	svc := NewURLService(repo)

	resp, err := svc.GetByID(context.Background(), "non-existent-id")

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, ErrURLNotFound, err)
}

func TestDelete(t *testing.T) {
	repo := newMockRepository()
	svc := NewURLService(repo)

	createReq := &models.CreateURLRequest{Original: "https://google.com"}
	created, _ := svc.Create(context.Background(), createReq)

	err := svc.Delete(context.Background(), created.ID)

	assert.NoError(t, err)

	_, err = svc.GetByID(context.Background(), created.ID)
	assert.Error(t, err)
}

func TestDeleteNotFound(t *testing.T) {
	repo := newMockRepository()
	svc := NewURLService(repo)

	err := svc.Delete(context.Background(), "non-existent-id")

	assert.Error(t, err)
	assert.Equal(t, ErrURLNotFound, err)
}

func TestList(t *testing.T) {
	repo := newMockRepository()
	svc := NewURLService(repo)

	for i := 0; i < 15; i++ {
		req := &models.CreateURLRequest{Original: "https://google.com"}
		_, err := svc.Create(context.Background(), req)
		assert.NoError(t, err)
	}

	resp, err := svc.List(context.Background(), 1, 10)

	assert.NoError(t, err)
	assert.Equal(t, 15, resp.Total)
	assert.Equal(t, 10, len(resp.URLs))
	assert.Equal(t, 2, resp.TotalPages)
}

func TestListPagination(t *testing.T) {
	repo := newMockRepository()
	svc := NewURLService(repo)

	for i := 0; i < 15; i++ {
		req := &models.CreateURLRequest{Original: "https://google.com"}
		_, err := svc.Create(context.Background(), req)
		assert.NoError(t, err)
	}

	resp, err := svc.List(context.Background(), 2, 10)

	assert.NoError(t, err)
	assert.Equal(t, 5, len(resp.URLs))
	assert.Equal(t, 2, resp.Page)
}