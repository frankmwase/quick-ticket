package payments

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"quick-ticket/domain"
)

// MWJsonProviderConfig holds the configuration for connecting to a
// Malawi Pay Standard compliant payment gateway.
type MWJsonProviderConfig struct {
	BaseURL    string
	APIKey     string
	MerchantID string
	Timeout    time.Duration
}

// MWJsonProvider implements domain.MonetizationExtension using the
// MW-JSON interoperable payment standard for Airtel Money, TNM Mpamba,
// and bank transfers within the Malawi financial ecosystem.
type MWJsonProvider struct {
	config MWJsonProviderConfig
	client *http.Client
}

func NewMWJsonProvider(config MWJsonProviderConfig) domain.MonetizationExtension {
	timeout := config.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}
	return &MWJsonProvider{
		config: config,
		client: &http.Client{Timeout: timeout},
	}
}

func (p *MWJsonProvider) VerifyAndProcessPayment(ctx context.Context, tx *domain.MWJsonTransaction) (bool, error) {
	// ── Structural Validation ────────────────────────────────────────────────
	if err := p.validateTransaction(tx); err != nil {
		return false, err
	}

	// ── Idempotency Guard ────────────────────────────────────────────────────
	if tx.Header.IdempotencyKey == "" {
		return false, errors.New("MW-JSON: idempotency key is required for payment processing")
	}

	// ── TTL Expiry Check ─────────────────────────────────────────────────────
	if tx.Header.TTL > 0 {
		expiresAt := tx.Header.Timestamp.Add(time.Duration(tx.Header.TTL) * time.Second)
		if time.Now().UTC().After(expiresAt) {
			return false, errors.New("MW-JSON: transaction TTL has expired")
		}
	}

	// ── Forward to Payment Gateway ───────────────────────────────────────────
	payload, err := json.Marshal(tx)
	if err != nil {
		return false, fmt.Errorf("MW-JSON: failed to marshal transaction: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.config.BaseURL+"/v1/transactions/process", bytes.NewReader(payload))
	if err != nil {
		return false, fmt.Errorf("MW-JSON: failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.config.APIKey)
	req.Header.Set("X-Merchant-ID", p.config.MerchantID)
	req.Header.Set("X-Idempotency-Key", tx.Header.IdempotencyKey)

	resp, err := p.client.Do(req)
	if err != nil {
		return false, fmt.Errorf("MW-JSON: gateway request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		var errBody struct {
			Error string `json:"error"`
		}
		_ = json.NewDecoder(resp.Body).Decode(&errBody)
		return false, &domain.ErrPaymentFailed{
			TransactionID: tx.Header.MsgID,
			Reason:        fmt.Sprintf("gateway returned %d: %s", resp.StatusCode, errBody.Error),
		}
	}

	var result struct {
		Success       bool   `json:"success"`
		TransactionID string `json:"transaction_id"`
		Status        string `json:"status"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false, fmt.Errorf("MW-JSON: failed to decode gateway response: %w", err)
	}

	return result.Success, nil
}

func (p *MWJsonProvider) validateTransaction(tx *domain.MWJsonTransaction) error {
	if tx.MWVersion == "" {
		return errors.New("MW-JSON: mwVersion is required")
	}
	if tx.Header.MsgID == "" {
		return errors.New("MW-JSON: header.msgId is required")
	}
	if tx.Payload.Currency != "MWK" {
		return errors.New("MW-JSON: currency must be MWK")
	}
	if tx.Payload.Amount <= 0 {
		return errors.New("MW-JSON: amount must be positive")
	}
	if tx.Payload.Sender.ID == "" {
		return errors.New("MW-JSON: sender.id is required")
	}
	if tx.Payload.Receiver.ID == "" {
		return errors.New("MW-JSON: receiver.id is required")
	}

	validTypes := map[string]bool{"P2M": true, "P2P": true}
	if !validTypes[tx.Payload.Type] {
		return fmt.Errorf("MW-JSON: invalid transaction type %q, must be P2M or P2P", tx.Payload.Type)
	}

	validIDTypes := map[string]bool{"MSISDN": true, "BANK_ACC": true, "MERCHANT_CODE": true}
	if !validIDTypes[tx.Payload.Sender.IDType] {
		return fmt.Errorf("MW-JSON: invalid sender idType %q", tx.Payload.Sender.IDType)
	}
	if !validIDTypes[tx.Payload.Receiver.IDType] {
		return fmt.Errorf("MW-JSON: invalid receiver idType %q", tx.Payload.Receiver.IDType)
	}

	return nil
}
