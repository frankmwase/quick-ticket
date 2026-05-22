package domain

import (
	"context"
	"time"
)

// WorkflowStatus represents the lifecycle state of a workflow pipeline.
type WorkflowStatus string

const (
	WorkflowPending   WorkflowStatus = "PENDING"
	WorkflowActive    WorkflowStatus = "ACTIVE"
	WorkflowCompleted WorkflowStatus = "COMPLETED"
	WorkflowFailed    WorkflowStatus = "FAILED"
	WorkflowCancelled WorkflowStatus = "CANCELLED"
)

// WorkflowStepType defines the kind of action a workflow step performs.
type WorkflowStepType string

const (
	StepGenerateTickets WorkflowStepType = "GENERATE_TICKETS"
	StepSendEmails      WorkflowStepType = "SEND_EMAILS"
	StepProcessPayment  WorkflowStepType = "PROCESS_PAYMENT"
	StepValidateGeo     WorkflowStepType = "VALIDATE_GEOFENCE"
	StepWebhookNotify   WorkflowStepType = "WEBHOOK_NOTIFY"
	StepRenderPDF       WorkflowStepType = "RENDER_PDF"
	StepCustom          WorkflowStepType = "CUSTOM"
)

// WorkflowStep represents a single step in a workflow pipeline.
type WorkflowStep struct {
	ID        string            `json:"id"`
	Type      WorkflowStepType  `json:"type"`
	Config    map[string]string `json:"config"`
	Order     int               `json:"order"`
	Status    WorkflowStatus    `json:"status"`
	Error     string            `json:"error,omitempty"`
	StartedAt *time.Time        `json:"started_at,omitempty"`
	EndedAt   *time.Time        `json:"ended_at,omitempty"`
}

// Workflow is the aggregate root for scheduled releases and multi-step pipelines.
// A workflow ties together a batch and a series of ordered steps that execute
// sequentially (e.g., generate → render → email → webhook).
type Workflow struct {
	ID           string
	TenantID     string
	BatchID      string
	Name         string
	Status       WorkflowStatus
	Steps        []WorkflowStep
	ScheduledAt  *time.Time
	ExecutedAt   *time.Time
	CompletedAt  *time.Time
	CreatedAt    time.Time
	RetryCount   int
	MaxRetries   int
}

// NextPendingStep returns the next step that hasn't been executed yet.
func (w *Workflow) NextPendingStep() *WorkflowStep {
	for i := range w.Steps {
		if w.Steps[i].Status == WorkflowPending {
			return &w.Steps[i]
		}
	}
	return nil
}

// IsScheduledNow returns true if the workflow is due for execution.
func (w *Workflow) IsScheduledNow() bool {
	if w.ScheduledAt == nil {
		return true // No schedule means execute immediately
	}
	return time.Now().UTC().After(*w.ScheduledAt)
}

// AllStepsCompleted returns true if every step has completed successfully.
func (w *Workflow) AllStepsCompleted() bool {
	for _, s := range w.Steps {
		if s.Status != WorkflowCompleted {
			return false
		}
	}
	return true
}

// ScheduledRelease holds the configuration for a time-delayed ticket release.
type ScheduledRelease struct {
	ID          string
	TenantID    string
	BatchID     string
	WorkflowID  string
	ReleaseAt   time.Time
	Executed    bool
	CreatedAt   time.Time
}

// WorkflowRepository defines the outbound port for workflow persistence.
type WorkflowRepository interface {
	Save(ctx context.Context, workflow *Workflow) error
	FindByID(ctx context.Context, id string) (*Workflow, error)
	FindPendingByTenant(ctx context.Context, tenantID string) ([]*Workflow, error)
	FindDueForExecution(ctx context.Context) ([]*Workflow, error)
	Update(ctx context.Context, workflow *Workflow) error
}

// WorkflowService defines the inbound port for workflow orchestration.
type WorkflowService interface {
	CreateWorkflow(ctx context.Context, workflow *Workflow) error
	ExecuteWorkflow(ctx context.Context, workflowID string) error
	CancelWorkflow(ctx context.Context, workflowID string) error
	ProcessScheduledReleases(ctx context.Context) error
	GetWorkflowStatus(ctx context.Context, workflowID string) (*Workflow, error)
}

// WebhookEvent represents an event dispatched to external listeners.
type WebhookEvent struct {
	Event     string      `json:"event"`
	TenantID  string      `json:"tenant_id"`
	Timestamp int64       `json:"timestamp"`
	Data      interface{} `json:"data"`
}

// WebhookDispatcher defines the outbound port for sending webhook callbacks.
type WebhookDispatcher interface {
	Dispatch(ctx context.Context, tenantID string, event WebhookEvent) error
}
