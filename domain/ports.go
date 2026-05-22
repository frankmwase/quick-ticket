package domain

import (
	"context"
)

type BulkGenerateRequest struct {
	TenantID   string
	BatchID    string
	Count      int
	OwnerID    string
	ManagedBy  string
	AutoEmail  bool
	TargetMail string
}

type VerificationContext struct {
	Token     string
	Latitude  float64
	Longitude float64
}

type VerificationResult struct {
	Valid    bool
	Message  string
	TicketID string
}

type TicketService interface {
	GenerateBulk(ctx context.Context, req BulkGenerateRequest) ([]*Ticket, error)
	VerifyTicket(ctx context.Context, vCtx VerificationContext) (*VerificationResult, error)
	UpdateStatus(ctx context.Context, ticketID string, nextStatus TicketStatus, actorID string, reason string) error
	DistributeTicket(ctx context.Context, ticketID string, targetEmail string) error
}

type TicketRepository interface {
	SaveBulk(ctx context.Context, tickets []*Ticket) error
	FindByID(ctx context.Context, id string) (*Ticket, error)
	FindByToken(ctx context.Context, token string) (*Ticket, error)
	Update(ctx context.Context, ticket *Ticket) error
}

type BatchRepository interface {
	FindByID(ctx context.Context, id string) (*TicketBatch, error)
	Save(ctx context.Context, batch *TicketBatch) error
}

type NotificationEngine interface {
	SendEmailWithAttachment(ctx context.Context, toEmail string, subject string, body string, pdfBytes []byte) error
}

type PDFRenderer interface {
	GenerateTicketPDF(ticket *Ticket, renderSpec []byte) ([]byte, error)
}
