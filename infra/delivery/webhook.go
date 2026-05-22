package delivery

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"quick-ticket/domain"
)

// WebhookConfig holds per-tenant webhook configuration.
type WebhookConfig struct {
	URL    string
	Secret string
}

// WebhookEngine implements domain.WebhookDispatcher.
// It dispatches signed JSON payloads to tenant-registered webhook URLs
// with retry logic and HMAC-SHA256 signature verification.
type WebhookEngine struct {
	client     *http.Client
	configFunc func(tenantID string) (*WebhookConfig, error)
}

// NewWebhookEngine creates a webhook dispatcher.
// configFunc should return the webhook URL and secret for a given tenant.
func NewWebhookEngine(configFunc func(tenantID string) (*WebhookConfig, error)) domain.WebhookDispatcher {
	return &WebhookEngine{
		client:     &http.Client{Timeout: 10 * time.Second},
		configFunc: configFunc,
	}
}

func (w *WebhookEngine) Dispatch(ctx context.Context, tenantID string, event domain.WebhookEvent) error {
	config, err := w.configFunc(tenantID)
	if err != nil {
		log.Printf("webhook: no config for tenant %s: %v", tenantID, err)
		return nil // Don't fail operations because a webhook isn't configured
	}

	if config.URL == "" {
		return nil // Webhooks not configured for this tenant
	}

	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("webhook: failed to marshal event: %w", err)
	}

	// Fire-and-forget with retries in a background goroutine
	go w.dispatchWithRetry(config, payload, 3)

	return nil
}

func (w *WebhookEngine) dispatchWithRetry(config *WebhookConfig, payload []byte, maxRetries int) {
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff: 1s, 2s, 4s
			time.Sleep(time.Duration(1<<uint(attempt-1)) * time.Second)
		}

		req, err := http.NewRequest(http.MethodPost, config.URL, bytes.NewReader(payload))
		if err != nil {
			log.Printf("webhook: failed to create request (attempt %d): %v", attempt, err)
			continue
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("User-Agent", "QuickTicket-Webhook/1.0")
		req.Header.Set("X-Webhook-Timestamp", fmt.Sprintf("%d", time.Now().Unix()))

		// HMAC-SHA256 signature for payload integrity verification
		if config.Secret != "" {
			signature := computeHMAC(payload, config.Secret)
			req.Header.Set("X-Webhook-Signature", signature)
		}

		resp, err := w.client.Do(req)
		if err != nil {
			log.Printf("webhook: delivery failed (attempt %d): %v", attempt, err)
			continue
		}
		resp.Body.Close()

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			log.Printf("webhook: delivered successfully to %s (attempt %d)", config.URL, attempt)
			return
		}

		log.Printf("webhook: received %d from %s (attempt %d)", resp.StatusCode, config.URL, attempt)
	}

	log.Printf("webhook: exhausted all %d retries for %s", maxRetries, config.URL)
}

func computeHMAC(payload []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	return hex.EncodeToString(mac.Sum(nil))
}
