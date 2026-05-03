package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"github.com/duiliofanton/go-shorten-url/internal/models"
	"github.com/duiliofanton/go-shorten-url/internal/repository"
	"github.com/duiliofanton/go-shorten-url/pkg/validator"
)

const (
	maxPerPage             = 100
	defaultPerPage         = 10
	shortCodeRetries       = 5
	shortCodeBytes         = 3
	idBytes                = 16
)

var (
	ErrInvalidURL    = errors.New("invalid URL format")
	ErrURLNotFound   = errors.New("url not found")
	ErrInternalError = errors.New("internal server error")
)

type URLService interface {
	Create(ctx context.Context, req *models.CreateURLRequest) (*models.URLResponse, error)
	GetByID(ctx context.Context, id string) (*models.URLResponse, error)
	GetByShortCode(ctx context.Context, shortCode string) (*models.URLResponse, error)
	Update(ctx context.Context, id string, req *models.UpdateURLRequest) (*models.URLResponse, error)
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, page, perPage int) (*models.ListURLsResponse, error)
}

type urlService struct {
	repo repository.URLRepository
}

func NewURLService(repo repository.URLRepository) URLService {
	return &urlService{repo: repo}
}

func (s *urlService) Create(ctx context.Context, req *models.CreateURLRequest) (*models.URLResponse, error) {
	if err := validator.ValidateURL(req.Original); err != nil {
		return nil, ErrInvalidURL
	}

	id, err := generateID()
	if err != nil {
		return nil, ErrInternalError
	}

	now := time.Now()
	for attempt := 0; attempt < shortCodeRetries; attempt++ {
		shortCode, err := generateShortCode()
		if err != nil {
			return nil, ErrInternalError
		}

		url := &models.URL{
			ID:        id,
			Original:  req.Original,
			ShortCode: shortCode,
			CreatedAt: now,
			UpdatedAt: now,
		}

		err = s.repo.Create(ctx, url)
		if err == nil {
			return toResponse(url), nil
		}
		if errors.Is(err, repository.ErrShortCodeConflict) {
			continue
		}
		return nil, ErrInternalError
	}
	return nil, ErrInternalError
}

func (s *urlService) GetByID(ctx context.Context, id string) (*models.URLResponse, error) {
	url, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrURLNotFound
	}
	return toResponse(url), nil
}

func (s *urlService) GetByShortCode(ctx context.Context, shortCode string) (*models.URLResponse, error) {
	url, err := s.repo.GetByShortCode(ctx, shortCode)
	if err != nil {
		return nil, ErrURLNotFound
	}
	return toResponse(url), nil
}

func (s *urlService) Update(ctx context.Context, id string, req *models.UpdateURLRequest) (*models.URLResponse, error) {
	if err := validator.ValidateURL(req.Original); err != nil {
		return nil, ErrInvalidURL
	}

	url, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrURLNotFound
	}

	url.Original = req.Original
	url.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, url); err != nil {
		return nil, ErrInternalError
	}
	return toResponse(url), nil
}

func (s *urlService) Delete(ctx context.Context, id string) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return ErrURLNotFound
	}
	return nil
}

func (s *urlService) List(ctx context.Context, page, perPage int) (*models.ListURLsResponse, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = defaultPerPage
	}
	if perPage > maxPerPage {
		perPage = maxPerPage
	}

	urls, total, err := s.repo.List(ctx, page, perPage)
	if err != nil {
		return nil, ErrInternalError
	}

	response := make([]models.URLResponse, len(urls))
	for i, url := range urls {
		response[i] = *toResponse(&url)
	}

	totalPages := (total + perPage - 1) / perPage
	if totalPages < 1 {
		totalPages = 1
	}

	return &models.ListURLsResponse{
		URLs:       response,
		Total:      total,
		Page:       page,
		PerPage:    perPage,
		TotalPages: totalPages,
	}, nil
}

func toResponse(url *models.URL) *models.URLResponse {
	return &models.URLResponse{
		ID:        url.ID,
		Original:  url.Original,
		ShortCode: url.ShortCode,
		CreatedAt: url.CreatedAt,
		UpdatedAt: url.UpdatedAt,
	}
}

func generateID() (string, error) {
	bytes := make([]byte, idBytes)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func generateShortCode() (string, error) {
	bytes := make([]byte, shortCodeBytes)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
