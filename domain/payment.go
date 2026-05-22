package domain

import (
	"context"
	"time"
)

type MWJsonHeader struct {
	MsgID          string    `json:"msgId"`
	Timestamp      time.Time `json:"timestamp"`
	TTL            int       `json:"ttl"`
	IdempotencyKey string    `json:"idempotencyKey"`
}

type MWJsonParticipant struct {
	ID     string `json:"id"`
	IDType string `json:"idType"`
}

type MWJsonPayload struct {
	Amount   float64           `json:"amount"`
	Currency string            `json:"currency"`
	Type     string            `json:"type"`
	Sender   MWJsonParticipant `json:"sender"`
	Receiver MWJsonParticipant `json:"receiver"`
}

type MWJsonTransaction struct {
	MWVersion string        `json:"mwVersion"`
	Header    MWJsonHeader  `json:"header"`
	Payload   MWJsonPayload `json:"payload"`
}

type MonetizationExtension interface {
	VerifyAndProcessPayment(ctx context.Context, tx *MWJsonTransaction) (bool, error)
}
