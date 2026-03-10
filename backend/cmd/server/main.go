package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/PhanAnh1001/ai-accounting/backend/internal/config"
	"github.com/PhanAnh1001/ai-accounting/backend/internal/handler"
	"github.com/PhanAnh1001/ai-accounting/backend/internal/middleware"
	"github.com/PhanAnh1001/ai-accounting/backend/internal/repository"
	"github.com/PhanAnh1001/ai-accounting/backend/internal/service"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	ctx := context.Background()
	pool, err := repository.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer pool.Close()

	// Repositories
	userRepo := repository.NewUserRepository(pool)
	orgRepo := repository.NewOrganizationRepository(pool)
	docRepo := repository.NewDocumentRepository(pool)
	entryRepo := repository.NewEntryRepository(pool)

	// Services
	var ocrSvc service.OCRService
	switch cfg.OCRProvider {
	case "mock":
		ocrSvc = &service.MockOCRService{Result: service.DefaultMockOCRResult()}
	case "google_vision":
		ocrSvc = service.NewGoogleVisionClient(cfg.OCRAPIKey)
	case "viettel":
		ocrSvc = service.NewViettelAIClient(cfg.OCRAPIKey)
	default: // "fpt"
		ocrSvc = service.NewFPTAIClient(cfg.OCRAPIKey)
	}
	ruleEngine := service.NewVASRuleEngine()
	processor := service.NewProcessorService(ocrSvc, ruleEngine, docRepo, entryRepo)
	if cfg.AnthropicAPIKey != "" {
		processor.WithClaudeEnricher(service.NewClaudeEnricher(cfg.AnthropicAPIKey))
	}

	// MISA adapter: use mock if no API URL configured
	var misaAdapter service.MISAAdapter
	if cfg.MISAApiURL != "" && cfg.MISAApiKey != "" {
		misaAdapter = service.NewMISAClient(cfg.MISAApiURL, cfg.MISAApiKey)
	} else {
		misaAdapter = &service.MockMISAAdapter{}
	}

	// MISA callback handler (public endpoint — MISA calls this, no JWT needed)
	misaCallbackHandler := handler.NewMISACallbackHandler(entryRepo, cfg.MISAAppID)

	// Handlers
	authHandler := handler.NewAuthHandler(userRepo, orgRepo, cfg.JWTSecret)
	documentHandler := handler.NewDocumentHandler(docRepo, cfg.UploadDir, processor)
	entryHandler := handler.NewEntryHandler(entryRepo, misaAdapter)
	settingsHandler := handler.NewSettingsHandler(orgRepo)
	exportHandler := handler.NewExportHandler(entryRepo)

	// Router
	r := chi.NewRouter()

	// Global middleware
	r.Use(chimw.Logger)
	r.Use(chimw.Recoverer)
	r.Use(chimw.RequestID)
	r.Use(chimw.RealIP)
	r.Use(chimw.Timeout(30 * time.Second))
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   cfg.CORSOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Public routes
	r.Get("/api/health", handler.Health)
	r.Post("/api/auth/register", authHandler.Register)
	r.Post("/api/auth/login", authHandler.Login)

	// MISA AMIS callback — public, protected by HMAC signature validation
	r.Post("/api/misa/callback", misaCallbackHandler.HandleCallback)

	// Protected auth routes
	r.Group(func(r chi.Router) {
		r.Use(middleware.Auth(cfg.JWTSecret))
		r.Post("/api/auth/change-password", authHandler.ChangePassword)
	})

	// Protected routes
	r.Group(func(r chi.Router) {
		r.Use(middleware.Auth(cfg.JWTSecret))

		r.Post("/api/documents", documentHandler.Upload)
		r.Get("/api/documents", documentHandler.List)
		r.Get("/api/documents/stats", documentHandler.Stats)
		r.Get("/api/documents/{id}", documentHandler.Get)
		r.Post("/api/documents/{id}/retry", documentHandler.Retry)

		r.Get("/api/entries", entryHandler.List)
		r.Get("/api/entries/stats", entryHandler.Stats)
		r.Get("/api/entries/export", exportHandler.ExportCSV)
		r.Post("/api/entries/bulk-approve", entryHandler.BulkApprove)
		r.Get("/api/entries/{id}", entryHandler.Get)
		r.Patch("/api/entries/{id}", entryHandler.Update)
		r.Post("/api/entries/{id}/approve", entryHandler.Approve)
		r.Post("/api/entries/{id}/reject", entryHandler.Reject)
		r.Post("/api/entries/{id}/sync", entryHandler.Sync)

		r.Get("/api/settings", settingsHandler.Get)
		r.Put("/api/settings", settingsHandler.Update)
	})

	// Server
	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	go func() {
		log.Printf("server starting on :%s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("shutting down server...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("server forced to shutdown: %v", err)
	}
	log.Println("server stopped")
}
