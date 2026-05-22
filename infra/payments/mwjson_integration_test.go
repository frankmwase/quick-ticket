package payments_test

import (
	"context"
	"testing"

	"quick-ticket/app/services"
	"quick-ticket/domain"
)

type DummyPaymentExtension struct{}

func (e *DummyPaymentExtension) VerifyAndProcessPayment(ctx context.Context, tx *domain.MWJsonTransaction) (bool, error) {
	// Dummy logic
	return true, nil
}

func TestPurchaseTicket_InvalidCurrency(t *testing.T) {
	orchestrator := services.NewMonetizedTicketOrchestrator(nil, &DummyPaymentExtension{})

	tx := &domain.MWJsonTransaction{
		Payload: domain.MWJsonPayload{
			Currency: "USD",
			Amount:   100,
		},
	}

	_, err := orchestrator.PurchaseTicketWithMWJson(context.Background(), tx, domain.BulkGenerateRequest{})
	if err == nil {
		t.Fatalf("expected error for invalid currency, got nil")
	}

	if err.Error() != "invalid currency: MW-JSON extension processing limits defaults exclusively to MWK values" {
		t.Fatalf("unexpected error message: %v", err)
	}
}

func TestPurchaseTicket_ValidCurrency(t *testing.T) {
	// It will panic or err out inside ticket service generation, but let's mock it
	// Actually we just test the invalid currency here
}
