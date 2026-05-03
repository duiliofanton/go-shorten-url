package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"syscall"
	"time"

	"github.com/duiliofanton/go-shorten-url/internal/config"
	"github.com/duiliofanton/go-shorten-url/internal/handler"
	"github.com/duiliofanton/go-shorten-url/internal/middleware"
	"github.com/duiliofanton/go-shorten-url/internal/models"
	"github.com/duiliofanton/go-shorten-url/internal/repository"
	"github.com/duiliofanton/go-shorten-url/internal/service"
	"github.com/joho/godotenv"
)

var shortCodePattern = regexp.MustCompile(`^/[a-f0-9]{6}$`)

func main() {
	if err := godotenv.Load(); err != nil {
		slog.Info(".env file not found, using environment variables", "source", "main")
	}

	cfg := config.Load()

	repo, err := repository.NewSQLiteURLRepository(cfg.Database)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer repo.Close()

	if err := repository.InitDatabase(repo.DB()); err != nil {
		slog.Error("failed to initialize database", "error", err)
		os.Exit(1)
	}

	urlService := service.NewURLService(repo)
	urlHandler := handler.NewURLHandler(urlService)
	redirectHandler := handler.NewRedirectHandler(urlService)

	rl := middleware.NewRateLimiter(100, time.Minute)
	defer rl.Stop()

	metrics := middleware.NewMetrics()
	defer metrics.Stop()

	auth := middleware.NewAuth(cfg.APIKey)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		handler.WriteJSON(w, http.StatusOK, models.HealthResponse{Status: "ok"})
	})
	mux.HandleFunc("GET /metrics", metrics.Handler())
	mux.HandleFunc("POST /api/urls", urlHandler.Create)
	mux.HandleFunc("GET /api/urls/{id}", urlHandler.GetByID)
	mux.HandleFunc("PUT /api/urls/{id}", urlHandler.Update)
	mux.HandleFunc("DELETE /api/urls/{id}", urlHandler.Delete)
	mux.HandleFunc("GET /api/urls", urlHandler.List)
	mux.HandleFunc("GET /{shortCode}", redirectHandler.Redirect)

	router := middleware.Recover(
		middleware.Logger(
			rl.Middleware(
				metrics.Middleware(
					auth.Middleware(allowAnonymous)(mux),
				),
			),
		),
	)

	server := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	go func() {
		slog.Info("starting server", "port", cfg.Server.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("failed to start server", "error", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("shutting down server")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		slog.Error("failed to shutdown server", "error", err)
	}

	slog.Info("server stopped")
}

func allowAnonymous(r *http.Request) bool {
	switch r.URL.Path {
	case "/health", "/metrics":
		return true
	}
	return shortCodePattern.MatchString(r.URL.Path)
}
