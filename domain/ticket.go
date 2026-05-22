package domain

import (
	"time"
)

type TicketStatus string

const (
	StatusPending   TicketStatus = "PENDING"
	StatusIssued    TicketStatus = "ISSUED"
	StatusValidated TicketStatus = "VALIDATED"
	StatusRevoked   TicketStatus = "REVOKED"
)

type StateEvent struct {
	Timestamp  time.Time    `json:"timestamp"`
	FromStatus TicketStatus `json:"from_status"`
	ToStatus   TicketStatus `json:"to_status"`
	ActorID    string       `json:"actor_id"`
	Reason     string       `json:"reason"`
}

type Ticket struct {
	ID          string
	TenantID    string
	BatchID     string
	OwnerID     string
	ManagedBy   string
	SecureToken string
	Status      TicketStatus
	History     []StateEvent
	CreatedAt   time.Time
}

type Geofence struct {
	Lat          float64 `json:"lat"`
	Lng          float64 `json:"lng"`
	RadiusMeters float64 `json:"radius_meters"`
}

type TicketBatch struct {
	ID               string
	TenantID         string
	Title            string
	ScheduledRelease *time.Time
	AutoSendEmail    bool
	Geofence         *Geofence
	RenderSpec       []byte
	CreatedAt        time.Time
}

func (b *TicketBatch) HasGeofence() bool {
	return b.Geofence != nil
}
