package main

import (
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"

	repo "github.com/tpt-nz/tpt-will-estate-nz/internal/repository"
	svc "github.com/tpt-nz/tpt-will-estate-nz/internal/services"
	"github.com/tpt-nz/tpt-will-estate-nz/internal/handlers"
	"github.com/tpt-nz/nz-common/middleware"
	"github.com/tpt-nz/realme-go"
	realmemw "github.com/tpt-nz/realme-go/middleware"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	// TODO: mirror config/env loading patterns from other apps.
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable"
	}

	pool, err := pgxpool.New(contextWithLogger(dbURL, logger))
	if err != nil {
		logger.Error("db connect", "err", err)
		os.Exit(1)
	}
	defer pool.Close()

	provider := realme.NewProviderFromEnv(logger) // TODO: implement helper usage based on other apps

	willRepo := repo.NewWillRepo(pool)

	encryptionSvc := svc.NewEncryptionService(logger)
	willSvc := svc.NewWillService(willRepo, encryptionSvc, logger)
	willHandler := handlers.NewWillHandler(willSvc, logger)

	r := chi.NewRouter()
	r.Use(middleware.RequestID())
	r.Use(middleware.Logger(logger))
	r.Use(middleware.RealMeContext(provider))
	r.Use(middleware.CorsFromEnv())

	// RealMe auth routes
	authHandler := handlers.NewAuthHandler(provider, logger)
	r.Route("/auth", func(rt chi.Router) {
		rt.Get("/login", authHandler.Login)
		rt.Get("/callback", authHandler.Callback)
		rt.Get("/logout", authHandler.Logout)
		rt.Get("/metadata", authHandler.Metadata)
		rt.Get("/status", authHandler.Status)
	})

	// Will routes
	r.Route("/api/wills", func(rt chi.Router) {
		rt.With(realmemw.RequireVerified(provider)).Post("", willHandler.Create)
		// more endpoints in later steps
	})

	srv := &http.Server{
		Addr:              ":8080",
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
	}

	logger.Info("listening", "addr", srv.Addr)
	if err := srv.ListenAndServe(); err != nil {
		logger.Error("server error", "err", err)
		os.Exit(1)
	}
}

// contextWithLogger creates a pgxpool config from URL.
// Scaffold only; other apps use proper config wiring.
func contextWithLogger(dbURL string, _ *slog.Logger) *pgxpool.Config {
	cfg, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		// best-effort fallback; will fail later with connection errors
		return &pgxpool.Config{}
	}
	return cfg
}


