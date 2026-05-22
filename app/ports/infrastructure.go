package ports

import (
	"context"

	"quick-ticket/domain"
)

// ───────────────────────────────────────────────────────────────────────────────
// OUTBOUND PORTS — Infrastructure isolation contracts
// ───────────────────────────────────────────────────────────────────────────────

// EmailSender abstracts email delivery (SMTP, SendGrid, Mailgun, etc.).
type EmailSender interface {
	domain.NotificationEngine
}

// SMSSender abstracts SMS delivery.
type SMSSender interface {
	SendSMS(ctx context.Context, toPhone string, body string) error
}

// PDFGenerator abstracts PDF rendering engines.
type PDFGenerator interface {
	domain.PDFRenderer
}

// QRCodeGenerator abstracts QR code generation.
type QRCodeGenerator interface {
	Generate(payload string, sizePx int) ([]byte, error)
}

// IdempotencyStore abstracts the idempotency key persistence layer.
type IdempotencyStore interface {
	Exists(ctx context.Context, key string) (bool, error)
	Get(ctx context.Context, key string) (statusCode int, body []byte, err error)
	Set(ctx context.Context, key string, statusCode int, body []byte, ttlSeconds int) error
}

// CacheStore abstracts general-purpose caching (Redis, in-memory, etc.).
type CacheStore interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, value []byte, ttlSeconds int) error
	Delete(ctx context.Context, key string) error
}

// EventBus abstracts internal event publishing for cross-service communication.
type EventBus interface {
	Publish(ctx context.Context, event domain.WebhookEvent) error
	Subscribe(eventType string, handler func(event domain.WebhookEvent)) error
}
