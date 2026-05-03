package repository

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/duiliofanton/go-shorten-url/internal/config"
	"github.com/duiliofanton/go-shorten-url/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestRepo(t *testing.T) *SQLiteURLRepository {
	t.Helper()
	cfg := config.DatabaseConfig{
		Dir:          t.TempDir(),
		Name:         "test.db",
		MaxOpenConns: 25,
		MaxIdleConns: 5,
		MaxLifetime:  5 * time.Minute,
	}
	repo, err := NewSQLiteURLRepository(cfg)
	require.NoError(t, err)
	t.Cleanup(repo.Close)
	require.NoError(t, InitDatabase(repo.DB()))
	return repo
}

func TestSQLiteRepository(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()
	now := time.Now()

	t.Run("Create and GetByID", func(t *testing.T) {
		url := &models.URL{
			ID:        "test-id-1",
			Original:  "https://example.com",
			ShortCode: "abc123",
			CreatedAt: now,
			UpdatedAt: now,
		}
		require.NoError(t, repo.Create(ctx, url))

		got, err := repo.GetByID(ctx, "test-id-1")
		require.NoError(t, err)
		assert.Equal(t, "test-id-1", got.ID)
		assert.Equal(t, "https://example.com", got.Original)
		assert.Equal(t, "abc123", got.ShortCode)
		assert.False(t, got.CreatedAt.IsZero())
		assert.WithinDuration(t, now, got.CreatedAt, time.Second)
	})

	t.Run("Create and GetByShortCode", func(t *testing.T) {
		url := &models.URL{
			ID:        "test-id-2",
			Original:  "https://google.com",
			ShortCode: "xyz789",
			CreatedAt: now,
			UpdatedAt: now,
		}
		require.NoError(t, repo.Create(ctx, url))

		got, err := repo.GetByShortCode(ctx, "xyz789")
		require.NoError(t, err)
		assert.Equal(t, "test-id-2", got.ID)
		assert.WithinDuration(t, now, got.UpdatedAt, time.Second)
	})

	t.Run("Update", func(t *testing.T) {
		url := &models.URL{
			ID:        "test-id-3",
			Original:  "https://old.com",
			ShortCode: "old123",
			CreatedAt: now,
			UpdatedAt: now,
		}
		require.NoError(t, repo.Create(ctx, url))

		url.Original = "https://new.com"
		url.UpdatedAt = now.Add(time.Hour)
		require.NoError(t, repo.Update(ctx, url))

		got, err := repo.GetByID(ctx, "test-id-3")
		require.NoError(t, err)
		assert.Equal(t, "https://new.com", got.Original)
	})

	t.Run("List", func(t *testing.T) {
		for i := 0; i < 5; i++ {
			s := strconv.Itoa(i)
			url := &models.URL{
				ID:        "test-list-" + s,
				Original:  "https://example.com/" + s,
				ShortCode: "lst" + s,
				CreatedAt: now.Add(time.Duration(i) * time.Hour),
				UpdatedAt: now.Add(time.Duration(i) * time.Hour),
			}
			require.NoError(t, repo.Create(ctx, url))
		}

		urls, total, err := repo.List(ctx, 1, 10)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, total, 5)
		assert.GreaterOrEqual(t, len(urls), 5)

		for _, url := range urls {
			assert.False(t, url.CreatedAt.IsZero(), "id %s", url.ID)
			assert.False(t, url.UpdatedAt.IsZero(), "id %s", url.ID)
		}
	})

	t.Run("List Pagination", func(t *testing.T) {
		urls, total, err := repo.List(ctx, 1, 2)
		require.NoError(t, err)
		assert.Equal(t, 8, total)
		assert.Len(t, urls, 2)
	})

	t.Run("Delete", func(t *testing.T) {
		require.NoError(t, repo.Delete(ctx, "test-id-1"))
		_, err := repo.GetByID(ctx, "test-id-1")
		assert.Error(t, err)
	})

	t.Run("GetByID Not Found", func(t *testing.T) {
		_, err := repo.GetByID(ctx, "non-existent-id")
		assert.Error(t, err)
	})

	t.Run("GetByShortCode Not Found", func(t *testing.T) {
		_, err := repo.GetByShortCode(ctx, "notfound")
		assert.Error(t, err)
	})

	t.Run("Create returns ErrShortCodeConflict on duplicate short code", func(t *testing.T) {
		first := &models.URL{
			ID:        "dup-1",
			Original:  "https://a.com",
			ShortCode: "duplic",
			CreatedAt: now,
			UpdatedAt: now,
		}
		require.NoError(t, repo.Create(ctx, first))

		second := &models.URL{
			ID:        "dup-2",
			Original:  "https://b.com",
			ShortCode: "duplic",
			CreatedAt: now,
			UpdatedAt: now,
		}
		err := repo.Create(ctx, second)
		assert.ErrorIs(t, err, ErrShortCodeConflict)
	})
}

func TestTimeRoundTrip(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()
	now := time.Now()

	for i := 0; i < 10; i++ {
		s := strconv.Itoa(i)
		url := &models.URL{
			ID:        "time-test-" + s,
			Original:  "https://test.com/" + s,
			ShortCode: "t" + s,
			CreatedAt: now.Add(time.Duration(i) * time.Hour),
			UpdatedAt: now.Add(time.Duration(i) * time.Hour),
		}
		require.NoError(t, repo.Create(ctx, url))
	}

	gotByID, err := repo.GetByID(ctx, "time-test-0")
	require.NoError(t, err)
	assert.WithinDuration(t, now, gotByID.CreatedAt, time.Second)

	gotByShortCode, err := repo.GetByShortCode(ctx, "t0")
	require.NoError(t, err)
	assert.WithinDuration(t, now, gotByShortCode.CreatedAt, time.Second)

	urls, total, err := repo.List(ctx, 1, 10)
	require.NoError(t, err)
	assert.Equal(t, 10, total)
	for _, url := range urls {
		assert.False(t, url.CreatedAt.IsZero(), "id %s", url.ID)
	}
}
