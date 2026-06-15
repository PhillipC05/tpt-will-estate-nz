package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/tpt-nz/tpt-will-estate-nz/internal/handlers"
	repo "github.com/tpt-nz/tpt-will-estate-nz/internal/repository"
	svc "github.com/tpt-nz/tpt-will-estate-nz/internal/services"
	nzmw "github.com/tpt-nz/nz-common/middleware"
	"github.com/tpt-nz/nz-common/linz"
	"github.com/tpt-nz/realme-go"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		// Development fallback — override DATABASE_URL in production.
		dbURL = "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable"
	}
	cfg, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		logger.Error("db config parse", "err", err)
		os.Exit(1)
	}
	pool, err := pgxpool.NewWithConfig(context.Background(), cfg)
	if err != nil {
		logger.Error("db connect", "err", err)
		os.Exit(1)
	}
	defer pool.Close()

	provider, err := realme.NewProvider(realmeConfigFromEnv())
	if err != nil {
		logger.Error("realme init", "err", err)
		os.Exit(1)
	}

	// Repositories
	willRepo := repo.NewWillRepo(pool)
	dalRepo := repo.NewDigitalAssetRepo(pool)
	auditRepo := repo.NewAuditRepo(pool)

	// Services
	encSvc := svc.NewEncryptionService(logger)
	auditSvc := svc.NewAuditService(auditRepo, logger)

	// Optional LINZ title verification — enabled when LINZ_API_KEY is set.
	var linzVerifier svc.PropertyVerifier
	if apiKey := os.Getenv("LINZ_API_KEY"); apiKey != "" {
		linzVerifier = svc.NewLINZPropertyVerifier(linz.NewClient(apiKey))
		logger.Info("LINZ title verification enabled")
	}

	willSvc := svc.NewWillService(willRepo, dalRepo, encSvc, auditSvc, linzVerifier, logger)
	deathTrigger := svc.NewDeathTrigger(willSvc, logger)

	// Handlers
	willHandler := handlers.NewWillHandler(willSvc, auditSvc, logger)
	authHandler := handlers.NewAuthHandler(provider, logger)
	executorHandler := handlers.NewExecutorHandler(willSvc, logger)
	beneficiaryHandler := handlers.NewBeneficiaryHandler(willSvc, logger)
	livenessHandler := handlers.NewLivenessHandler(willSvc, logger)

	corsOrigins := strings.Fields(os.Getenv("CORS_ORIGINS"))
	if len(corsOrigins) == 0 {
		corsOrigins = []string{"http://localhost:3000"}
	}

	r := chi.NewRouter()
	r.Use(nzmw.RequestID)
	r.Use(nzmw.Logger(logger))
	r.Use(nzmw.CORS(nzmw.CORSConfig{AllowedOrigins: corsOrigins, AllowCredentials: true}))

	// RealMe auth routes
	r.Route("/auth", func(rt chi.Router) {
		rt.Get("/login", authHandler.Login)
		rt.Get("/callback", authHandler.Callback)
		rt.Get("/logout", authHandler.Logout)
		rt.Get("/metadata", authHandler.Metadata)
		rt.Get("/status", authHandler.Status)
	})

	// Will routes — all require RealMe Verified Identity
	r.Route("/api/wills", func(rt chi.Router) {
		rt.With(provider.RequireVerified()).Post("", willHandler.Create)
		rt.With(provider.RequireVerified()).Get("/{id}", willHandler.GetByID)
		rt.With(provider.RequireVerified()).Post("/{id}/sign/testator", willHandler.SignTestator)
		rt.With(provider.RequireVerified()).Post("/{id}/sign/witness", willHandler.SignWitness)
		rt.With(provider.RequireVerified()).Post("/{id}/lock", willHandler.Lock)

		// Feature 4: Codicil
		rt.With(provider.RequireVerified()).Post("/{id}/codicil", willHandler.CreateCodicil)

		// Feature 2: Testator liveness confirmation (self-serve)
		rt.With(provider.RequireVerified()).Post("/{id}/confirm", willHandler.ConfirmLiveness)

		// Feature 7: Digital asset clauses
		rt.With(provider.RequireVerified()).Post("/{id}/digital-asset-clauses", willHandler.AddDigitalAssetClause)

		// Feature 8: LINZ property cross-reference
		rt.With(provider.RequireVerified()).Post("/{id}/property-clauses", willHandler.AddPropertyClause)

		// Feature 6: Hash-chained audit trail
		rt.With(provider.RequireVerified()).Get("/{id}/audit", willHandler.GetAuditLog)

		// Feature 1 + 5: Executor access (FLT verification + time-lock lazy unlock)
		rt.With(provider.RequireVerified()).Get("/{id}/executor", executorHandler.GetWill)

		rt.With(provider.RequireVerified()).Post("/{id}/beneficiaries/notify", beneficiaryHandler.NotifyAll)
	})

	// Feature 2: Liveness reminder — called by scheduler, not end users.
	// Restrict to internal network or use a separate LISTEN_ADDR_INTERNAL in production.
	r.Route("/internal/liveness", func(rt chi.Router) {
		rt.Get("/check", livenessHandler.BulkCheckLiveness)
		rt.Post("/wills/{id}/remind", livenessHandler.SendReminder)
	})

	// BDM death notification webhook (HMAC-SHA256 signed by BDM)
	r.Post("/webhooks/bdm/death", deathTrigger.HandleBDMNotification)

	addr := os.Getenv("LISTEN_ADDR")
	if addr == "" {
		addr = ":8080"
	}
	srv := &http.Server{
		Addr:              addr,
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
	}

	logger.Info("listening", "addr", srv.Addr)
	if err := srv.ListenAndServe(); err != nil {
		logger.Error("server error", "err", err)
		os.Exit(1)
	}
}

// realmeConfigFromEnv builds a realme.Config from environment variables.
func realmeConfigFromEnv() realme.Config {
	env := realme.Environment(os.Getenv("REALME_ENVIRONMENT"))
	if env == "" {
		env = realme.MTS
	}
	return realme.Config{
		Environment:    env,
		EntityID:       os.Getenv("REALME_ENTITY_ID"),
		ACSURL:         os.Getenv("REALME_ACS_URL"),
		CertFile:       os.Getenv("REALME_CERT_FILE"),
		KeyFile:        os.Getenv("REALME_KEY_FILE"),
		IdPMetadataURL: os.Getenv("REALME_IDP_METADATA_URL"),
	}
}
