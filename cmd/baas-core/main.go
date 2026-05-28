package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"quick-ticket/app/services"
	"quick-ticket/infra/db"
	"quick-ticket/infra/delivery"
	"quick-ticket/infra/payments"
	"quick-ticket/infra/rendering"
	httpPort "quick-ticket/interfaces/http"
	"quick-ticket/interfaces/http/handlers"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("Starting Quick Ticket BaaS...")

	// ── Configuration ────────────────────────────────────────────────────────
	dbURL := envOrDefault("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/quickticket?sslmode=disable")
	port := envOrDefault("PORT", "8080")
	smtpHost := envOrDefault("SMTP_HOST", "localhost")
	smtpPort := envOrDefault("SMTP_PORT", "587")
	smtpUser := envOrDefault("SMTP_USER", "")
	smtpPass := envOrDefault("SMTP_PASS", "")
	smtpFrom := envOrDefault("SMTP_FROM", "tickets@quickticket.io")
	mwjsonURL := envOrDefault("MWJSON_BASE_URL", "https://api.pay.mw")
	mwjsonKey := envOrDefault("MWJSON_API_KEY", "")
	mwjsonMerchant := envOrDefault("MWJSON_MERCHANT_ID", "")

	// ── Database Connection Pool ─────────────────────────────────────────────
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	log.Println("Database connection established")

	// ── Repository Layer (Outbound Ports) ────────────────────────────────────
	ticketRepo := db.NewPostgresTicketRepo(pool)
	batchRepo := db.NewPostgresBatchRepo(pool)
	workflowRepo := db.NewPostgresWorkflowRepo(pool)
	tenantRepo := db.NewPostgresTenantRepo(pool)

	// ── Infrastructure Services ──────────────────────────────────────────────
	smtpPortInt := 587
	if smtpPort != "" {
		if p, err := strconv.Atoi(smtpPort); err == nil {
			smtpPortInt = p
		}
	}

	notifier := delivery.NewSMTPNotifier(delivery.SMTPConfig{
		Host:     smtpHost,
		Port:     smtpPortInt,
		Username: smtpUser,
		Password: smtpPass,
		From:     smtpFrom,
		FromName: "Quick Ticket",
	})

	pdfEngine := rendering.NewFPDFRenderer()

	paymentProvider := payments.NewMWJsonProvider(payments.MWJsonProviderConfig{
		BaseURL:    mwjsonURL,
		APIKey:     mwjsonKey,
		MerchantID: mwjsonMerchant,
		Timeout:    30 * time.Second,
	})

	webhookEngine := delivery.NewWebhookEngine(func(tenantID string) (*delivery.WebhookConfig, error) {
		tenant, err := tenantRepo.FindByID(tenantID)
		if err != nil {
			return nil, err
		}
		return &delivery.WebhookConfig{
			URL:    tenant.Settings.WebhookURL,
			Secret: tenant.Settings.WebhookSecret,
		}, nil
	})

	// ── Service Layer (Inbound Ports) ────────────────────────────────────────
	ticketService := services.NewTicketService(ticketRepo, batchRepo, pdfEngine, notifier, services.WithWebhooks(webhookEngine))
	workflowService := services.NewWorkflowService(workflowRepo, batchRepo, ticketService, webhookEngine, notifier, pdfEngine)
	paymentOrchestrator := services.NewMonetizedTicketOrchestrator(ticketService, paymentProvider)

	// ── HTTP Handlers ────────────────────────────────────────────────────────
	ticketHandler := handlers.NewTicketHandler(ticketService)
	batchHandler := handlers.NewBatchHandler(batchRepo)
	workflowHandler := handlers.NewWorkflowHandler(workflowService)
	extHandler := handlers.NewExtensionHandler(paymentOrchestrator)
	healthHandler := handlers.NewHealthHandler()
	profileHandler := handlers.NewProfileHandler(pool)

	// ── Router ───────────────────────────────────────────────────────────────
	router := httpPort.SetupRouter(ticketHandler, batchHandler, workflowHandler, extHandler, healthHandler, profileHandler, tenantRepo)

	// ── Background Workers ───────────────────────────────────────────────────
	go startSchedulerWorker(workflowService)

	// ── HTTP Server ──────────────────────────────────────────────────────────
	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh

		log.Println("Shutting down gracefully...")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := srv.Shutdown(shutdownCtx); err != nil {
			log.Fatalf("Server shutdown failed: %v", err)
		}
	}()

	log.Printf("Quick Ticket BaaS listening on :%s", port)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server failed: %v", err)
	}

	log.Println("Server stopped")
}

// startSchedulerWorker runs a background loop that processes due workflows
// every 30 seconds.
func startSchedulerWorker(ws interface{ ProcessScheduledReleases(ctx context.Context) error }) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		if err := ws.ProcessScheduledReleases(ctx); err != nil {
			log.Printf("Scheduler worker error: %v", err)
		}
		cancel()
	}
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
