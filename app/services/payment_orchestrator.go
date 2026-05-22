package services

import (
	"context"
	"errors"
	"quick-ticket/domain"
)

type MonetizedTicketOrchestrator struct {
	ticketService   domain.TicketService
	paymentProvider domain.MonetizationExtension
}

func NewMonetizedTicketOrchestrator(ts domain.TicketService, mp domain.MonetizationExtension) *MonetizedTicketOrchestrator {
	return &MonetizedTicketOrchestrator{ticketService: ts, paymentProvider: mp}
}

func (o *MonetizedTicketOrchestrator) PurchaseTicketWithMWJson(ctx context.Context, tx *domain.MWJsonTransaction, req domain.BulkGenerateRequest) ([]*domain.Ticket, error) {
	if tx.Payload.Currency != "MWK" {
		return nil, errors.New("invalid currency: MW-JSON extension processing limits defaults exclusively to MWK values")
	}

	success, err := o.paymentProvider.VerifyAndProcessPayment(ctx, tx)
	if err != nil || !success {
		return nil, errors.New("transaction rejected across standard financial distribution clearance networks")
	}

	return o.ticketService.GenerateBulk(ctx, req)
}
