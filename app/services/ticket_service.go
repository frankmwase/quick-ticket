package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"math"
	"time"

	"quick-ticket/domain"

	"github.com/google/uuid"
)

type ticketService struct {
	ticketRepo domain.TicketRepository
	batchRepo  domain.BatchRepository
	pdfEngine  domain.PDFRenderer
	notifier   domain.NotificationEngine
	webhooks   domain.WebhookDispatcher
}

func NewTicketService(tr domain.TicketRepository, br domain.BatchRepository, pe domain.PDFRenderer, ne domain.NotificationEngine, opts ...TicketServiceOption) domain.TicketService {
	s := &ticketService{ticketRepo: tr, batchRepo: br, pdfEngine: pe, notifier: ne}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// TicketServiceOption allows optional dependencies to be injected.
type TicketServiceOption func(*ticketService)

// WithWebhooks injects a webhook dispatcher into the ticket service.
func WithWebhooks(wd domain.WebhookDispatcher) TicketServiceOption {
	return func(s *ticketService) {
		s.webhooks = wd
	}
}

func (s *ticketService) GenerateBulk(ctx context.Context, req domain.BulkGenerateRequest) ([]*domain.Ticket, error) {
	_, err := s.batchRepo.FindByID(ctx, req.BatchID)
	if err != nil {
		return nil, fmt.Errorf("invalid processing batch ID specified: %w", err)
	}

	tickets := make([]*domain.Ticket, req.Count)
	for i := 0; i < req.Count; i++ {
		token, _ := generateSecureCryptographicToken()
		tickets[i] = &domain.Ticket{
			ID:          uuid.New().String(),
			TenantID:    req.TenantID,
			BatchID:     req.BatchID,
			OwnerID:     req.OwnerID,
			ManagedBy:   req.ManagedBy,
			SecureToken: token,
			Status:      domain.StatusIssued,
			CreatedAt:   time.Now().UTC(),
		}
	}

	if err := s.ticketRepo.SaveBulk(ctx, tickets); err != nil {
		return nil, fmt.Errorf("bulk insertion subsystem failure: %w", err)
	}

	// Dispatch webhook event for bulk generation
	s.dispatchEvent(ctx, req.TenantID, "ticket.batch_generated", map[string]interface{}{
		"batch_id": req.BatchID,
		"count":    req.Count,
	})

	if req.AutoEmail && req.TargetMail != "" {
		go func() {
			_ = s.DistributeTicket(context.Background(), tickets[0].ID, req.TargetMail)
		}()
	}

	return tickets, nil
}

func generateSecureCryptographicToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func calculateHaversineDistance(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371e3 // Earth radius in meters
	phi1 := lat1 * math.Pi / 180
	phi2 := lat2 * math.Pi / 180
	deltaPhi := (lat2 - lat1) * math.Pi / 180
	deltaLambda := (lon2 - lon1) * math.Pi / 180

	a := math.Sin(deltaPhi/2)*math.Sin(deltaPhi/2) +
		math.Cos(phi1)*math.Cos(phi2)*
			math.Sin(deltaLambda/2)*math.Sin(deltaLambda/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return R * c
}

func (s *ticketService) VerifyTicket(ctx context.Context, vCtx domain.VerificationContext) (*domain.VerificationResult, error) {
	ticket, err := s.ticketRepo.FindByToken(ctx, vCtx.Token)
	if err != nil {
		return &domain.VerificationResult{Valid: false, Message: "Identity token not found"}, nil
	}

	batch, _ := s.batchRepo.FindByID(ctx, ticket.BatchID)

	if batch != nil && batch.HasGeofence() {
		distance := calculateHaversineDistance(vCtx.Latitude, vCtx.Longitude, batch.Geofence.Lat, batch.Geofence.Lng)
		if distance > batch.Geofence.RadiusMeters {
			return &domain.VerificationResult{
				Valid:   false,
				Message: fmt.Sprintf("Validation rejected: out of geographic boundaries. Distance: %f meters", distance),
			}, nil
		}
	}

	if ticket.Status != domain.StatusIssued {
		return &domain.VerificationResult{Valid: false, Message: "Ticket is no longer valid or has been revoked"}, nil
	}

	return &domain.VerificationResult{Valid: true, TicketID: ticket.ID}, nil
}

func (s *ticketService) UpdateStatus(ctx context.Context, ticketID string, nextStatus domain.TicketStatus, actorID string, reason string) error {
	ticket, err := s.ticketRepo.FindByID(ctx, ticketID)
	if err != nil {
		return err
	}

	prevStatus := ticket.Status

	event := domain.StateEvent{
		Timestamp:  time.Now().UTC(),
		FromStatus: prevStatus,
		ToStatus:   nextStatus,
		ActorID:    actorID,
		Reason:     reason,
	}

	ticket.Status = nextStatus
	ticket.History = append(ticket.History, event)

	if err := s.ticketRepo.Update(ctx, ticket); err != nil {
		return err
	}

	// Dispatch webhook for every status transition
	eventName := "ticket." + string(nextStatus)
	s.dispatchEvent(ctx, ticket.TenantID, eventName, map[string]interface{}{
		"ticket_id":   ticket.ID,
		"from_status": string(prevStatus),
		"to_status":   string(nextStatus),
		"actor_id":    actorID,
		"reason":      reason,
		"occurred_at": time.Now().UTC(),
	})

	return nil
}

func (s *ticketService) DistributeTicket(ctx context.Context, ticketID string, targetEmail string) error {
	ticket, err := s.ticketRepo.FindByID(ctx, ticketID)
	if err != nil {
		return err
	}

	batch, err := s.batchRepo.FindByID(ctx, ticket.BatchID)
	if err != nil {
		return err
	}

	var renderSpec []byte
	if batch != nil {
		renderSpec = batch.RenderSpec
	}

	pdfBytes, err := s.pdfEngine.GenerateTicketPDF(ticket, renderSpec)
	if err != nil {
		return err
	}

	return s.notifier.SendEmailWithAttachment(ctx, targetEmail, "Your Ticket", "Please find your ticket attached.", pdfBytes)
}

// dispatchEvent fires a webhook event if a dispatcher is configured.
// Failures are logged but never propagated — webhook delivery must not
// block core ticket operations.
func (s *ticketService) dispatchEvent(ctx context.Context, tenantID string, eventName string, data interface{}) {
	if s.webhooks == nil {
		return
	}
	event := domain.WebhookEvent{
		Event:     eventName,
		TenantID:  tenantID,
		Timestamp: time.Now().Unix(),
		Data:      data,
	}
	if err := s.webhooks.Dispatch(ctx, tenantID, event); err != nil {
		log.Printf("webhook dispatch failed for event %q: %v", eventName, err)
	}
}
