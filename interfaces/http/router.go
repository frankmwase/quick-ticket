package http

import (
	"quick-ticket/domain"
	"quick-ticket/interfaces/http/handlers"
	"quick-ticket/interfaces/http/middleware"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
)

func SetupRouter(
	ticketHandler *handlers.TicketHandler,
	batchHandler *handlers.BatchHandler,
	workflowHandler *handlers.WorkflowHandler,
	extHandler *handlers.ExtensionHandler,
	healthHandler *handlers.HealthHandler,
	profileHandler *handlers.ProfileHandler,
	tenantRepo domain.TenantRepository,
) *chi.Mux {
	r := chi.NewRouter()

	// ── Global Middleware ─────────────────────────────────────────────────────
	r.Use(chiMiddleware.RequestID)
	r.Use(chiMiddleware.RealIP)
	r.Use(chiMiddleware.Logger)
	r.Use(chiMiddleware.Recoverer)
	r.Use(middleware.CORS)

	// ── Health (unauthenticated) ─────────────────────────────────────────────
	r.Get("/health", healthHandler.Health)

	// ── API v1 (authenticated, tenant-scoped) ────────────────────────────────
	r.Route("/api/v1", func(r chi.Router) {
		// Apply auth and tenant scoping to all v1 routes
		if tenantRepo != nil {
			r.Use(middleware.Auth(tenantRepo))
			r.Use(middleware.TenantScope)
		}

		// Batch endpoints
		r.Route("/batches", func(r chi.Router) {
			r.Post("/", batchHandler.CreateBatch)
			r.Get("/{id}/status", batchHandler.GetBatchStatus)
		})

		// Core ticket endpoints
		r.Route("/tickets", func(r chi.Router) {
			r.With(middleware.Idempotency).Post("/generate", ticketHandler.GenerateTickets)
			r.Post("/verify", ticketHandler.VerifyTicket)
			r.Post("/revoke", ticketHandler.RevokeTicket)
			r.Patch("/{id}/status", ticketHandler.UpdateStatus)
		})

		// Workflow endpoints
		r.Route("/workflows", func(r chi.Router) {
			r.Post("/", workflowHandler.CreateWorkflow)
			r.Post("/{id}/execute", workflowHandler.ExecuteWorkflow)
			r.Post("/{id}/cancel", workflowHandler.CancelWorkflow)
			r.Get("/{id}/status", workflowHandler.GetWorkflowStatus)
		})

		// Extension endpoints (MW-JSON checkout)
		r.Route("/extensions", func(r chi.Router) {
			r.With(middleware.Idempotency).Post("/checkout", extHandler.Checkout)
		})

		// Profile and Member endpoints
		r.Route("/profile", func(r chi.Router) {
			r.Get("/", profileHandler.GetProfile)
			r.Put("/", profileHandler.UpdateProfile)
		})

		r.Route("/members", func(r chi.Router) {
			r.Get("/", profileHandler.GetMembers)
			r.Post("/", profileHandler.CreateMember)
		})
	})

	return r
}
