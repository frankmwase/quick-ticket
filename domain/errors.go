package domain

import "fmt"

// ErrNotFound indicates a requested entity does not exist.
type ErrNotFound struct {
	Entity string
	ID     string
}

func (e *ErrNotFound) Error() string {
	return fmt.Sprintf("%s with ID %q not found", e.Entity, e.ID)
}

// ErrInvalidTransition indicates an illegal ticket status transition was attempted.
type ErrInvalidTransition struct {
	From TicketStatus
	To   TicketStatus
}

func (e *ErrInvalidTransition) Error() string {
	return fmt.Sprintf("invalid status transition from %s to %s", e.From, e.To)
}

// ErrUnauthorized indicates the caller lacks permission for the requested operation.
type ErrUnauthorized struct {
	TenantID string
	Reason   string
}

func (e *ErrUnauthorized) Error() string {
	return fmt.Sprintf("unauthorized for tenant %q: %s", e.TenantID, e.Reason)
}

// ErrGeofenceViolation indicates a ticket operation was attempted outside the allowed geographic boundary.
type ErrGeofenceViolation struct {
	DistanceMeters float64
	RadiusMeters   float64
}

func (e *ErrGeofenceViolation) Error() string {
	return fmt.Sprintf("geofence violation: distance %.2fm exceeds allowed radius %.2fm", e.DistanceMeters, e.RadiusMeters)
}

// ErrDuplicateIdempotencyKey indicates a replay of an already-processed request.
type ErrDuplicateIdempotencyKey struct {
	Key string
}

func (e *ErrDuplicateIdempotencyKey) Error() string {
	return fmt.Sprintf("duplicate idempotency key: %s", e.Key)
}

// ErrPaymentFailed indicates a payment processing failure via the MW-JSON bridge.
type ErrPaymentFailed struct {
	TransactionID string
	Reason        string
}

func (e *ErrPaymentFailed) Error() string {
	return fmt.Sprintf("payment failed for transaction %q: %s", e.TransactionID, e.Reason)
}

// ErrBatchNotReady indicates a batch is not yet released or is past its release window.
type ErrBatchNotReady struct {
	BatchID string
	Reason  string
}

func (e *ErrBatchNotReady) Error() string {
	return fmt.Sprintf("batch %q not ready: %s", e.BatchID, e.Reason)
}

// ErrValidation indicates a generic input validation failure.
type ErrValidation struct {
	Field   string
	Message string
}

func (e *ErrValidation) Error() string {
	return fmt.Sprintf("validation error on field %q: %s", e.Field, e.Message)
}
