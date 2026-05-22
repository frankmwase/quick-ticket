// Package ports defines inbound/outbound contract interfaces for the application layer.
// These are pure Go interfaces with no framework or infrastructure dependencies.
package ports

import (
	"context"

	"quick-ticket/domain"
)

// ───────────────────────────────────────────────────────────────────────────────
// INBOUND PORTS — Use-case / Service layer contracts
// ───────────────────────────────────────────────────────────────────────────────

// TicketUseCase exposes the core ticket operations to delivery adapters.
type TicketUseCase interface {
	domain.TicketService
}

// WorkflowUseCase exposes the workflow/scheduling operations.
type WorkflowUseCase interface {
	domain.WorkflowService
}

// TenantUseCase exposes tenant management operations.
type TenantUseCase interface {
	CreateTenant(ctx context.Context, tenant *domain.Tenant) error
	GetTenant(ctx context.Context, id string) (*domain.Tenant, error)
	UpdateSettings(ctx context.Context, id string, settings domain.TenantSettings) error
	RotateAPIKey(ctx context.Context, id string) (newKey string, err error)
}

// CheckoutUseCase exposes the monetised ticket purchase flow.
type CheckoutUseCase interface {
	PurchaseTicketWithMWJson(ctx context.Context, tx *domain.MWJsonTransaction, req domain.BulkGenerateRequest) ([]*domain.Ticket, error)
}
