package repository

import (
	"context"

	"github.com/duiliofanton/go-shorten-url/internal/models"
)

type URLRepository interface {
	Create(ctx context.Context, url *models.URL) error
	GetByID(ctx context.Context, id string) (*models.URL, error)
	GetByShortCode(ctx context.Context, shortCode string) (*models.URL, error)
	Update(ctx context.Context, url *models.URL) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, page, perPage int) ([]models.URL, int, error)
}