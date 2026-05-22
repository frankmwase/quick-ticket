package services

import (
	"context"
	"fmt"
	"log"
	"time"

	"quick-ticket/domain"
)

type workflowService struct {
	workflowRepo domain.WorkflowRepository
	batchRepo    domain.BatchRepository
	ticketSvc    domain.TicketService
	webhooks     domain.WebhookDispatcher
	notifier     domain.NotificationEngine
	pdfEngine    domain.PDFRenderer
}

// NewWorkflowService constructs a workflow service with all required dependencies
// injected through interfaces.
func NewWorkflowService(
	wr domain.WorkflowRepository,
	br domain.BatchRepository,
	ts domain.TicketService,
	wh domain.WebhookDispatcher,
	ne domain.NotificationEngine,
	pe domain.PDFRenderer,
) domain.WorkflowService {
	return &workflowService{
		workflowRepo: wr,
		batchRepo:    br,
		ticketSvc:    ts,
		webhooks:     wh,
		notifier:     ne,
		pdfEngine:    pe,
	}
}

func (s *workflowService) CreateWorkflow(ctx context.Context, workflow *domain.Workflow) error {
	if workflow.Name == "" {
		return &domain.ErrValidation{Field: "name", Message: "workflow name is required"}
	}

	workflow.Status = domain.WorkflowPending
	workflow.CreatedAt = time.Now().UTC()

	// Initialise all steps as pending
	for i := range workflow.Steps {
		workflow.Steps[i].Status = domain.WorkflowPending
	}

	return s.workflowRepo.Save(ctx, workflow)
}

func (s *workflowService) ExecuteWorkflow(ctx context.Context, workflowID string) error {
	workflow, err := s.workflowRepo.FindByID(ctx, workflowID)
	if err != nil {
		return err
	}

	if workflow.Status == domain.WorkflowCompleted || workflow.Status == domain.WorkflowCancelled {
		return &domain.ErrValidation{Field: "status", Message: "workflow already completed or cancelled"}
	}

	if !workflow.IsScheduledNow() {
		return &domain.ErrBatchNotReady{BatchID: workflow.BatchID, Reason: "scheduled release time has not arrived"}
	}

	now := time.Now().UTC()
	workflow.Status = domain.WorkflowActive
	workflow.ExecutedAt = &now

	for {
		step := workflow.NextPendingStep()
		if step == nil {
			break
		}

		stepStart := time.Now().UTC()
		step.StartedAt = &stepStart

		err := s.executeStep(ctx, workflow, step)
		stepEnd := time.Now().UTC()
		step.EndedAt = &stepEnd

		if err != nil {
			step.Status = domain.WorkflowFailed
			step.Error = err.Error()
			workflow.Status = domain.WorkflowFailed

			if saveErr := s.workflowRepo.Update(ctx, workflow); saveErr != nil {
				log.Printf("failed to save workflow state after step failure: %v", saveErr)
			}

			return fmt.Errorf("workflow step %q (%s) failed: %w", step.ID, step.Type, err)
		}

		step.Status = domain.WorkflowCompleted
	}

	completedAt := time.Now().UTC()
	workflow.Status = domain.WorkflowCompleted
	workflow.CompletedAt = &completedAt

	return s.workflowRepo.Update(ctx, workflow)
}

func (s *workflowService) executeStep(ctx context.Context, workflow *domain.Workflow, step *domain.WorkflowStep) error {
	switch step.Type {
	case domain.StepGenerateTickets:
		count := 1
		if v, ok := step.Config["count"]; ok {
			fmt.Sscanf(v, "%d", &count)
		}
		_, err := s.ticketSvc.GenerateBulk(ctx, domain.BulkGenerateRequest{
			TenantID: workflow.TenantID,
			BatchID:  workflow.BatchID,
			Count:    count,
			OwnerID:  step.Config["owner_id"],
		})
		return err

	case domain.StepSendEmails:
		// Email distribution is handled per-ticket by the ticket service
		targetEmail := step.Config["target_email"]
		ticketID := step.Config["ticket_id"]
		if targetEmail != "" && ticketID != "" {
			return s.ticketSvc.DistributeTicket(ctx, ticketID, targetEmail)
		}
		return nil

	case domain.StepWebhookNotify:
		if s.webhooks == nil {
			return nil
		}
		return s.webhooks.Dispatch(ctx, workflow.TenantID, domain.WebhookEvent{
			Event:     "workflow.step_completed",
			TenantID:  workflow.TenantID,
			Timestamp: time.Now().Unix(),
			Data: map[string]string{
				"workflow_id": workflow.ID,
				"step_id":     step.ID,
				"step_type":   string(step.Type),
			},
		})

	case domain.StepCustom:
		// Custom steps are extensibility hooks — the config map carries
		// arbitrary instructions for third-party consumers.
		log.Printf("executing custom workflow step %q with config: %v", step.ID, step.Config)
		return nil

	default:
		return fmt.Errorf("unknown workflow step type: %s", step.Type)
	}
}

func (s *workflowService) CancelWorkflow(ctx context.Context, workflowID string) error {
	workflow, err := s.workflowRepo.FindByID(ctx, workflowID)
	if err != nil {
		return err
	}

	if workflow.Status == domain.WorkflowCompleted {
		return &domain.ErrValidation{Field: "status", Message: "cannot cancel a completed workflow"}
	}

	workflow.Status = domain.WorkflowCancelled
	return s.workflowRepo.Update(ctx, workflow)
}

func (s *workflowService) ProcessScheduledReleases(ctx context.Context) error {
	workflows, err := s.workflowRepo.FindDueForExecution(ctx)
	if err != nil {
		return err
	}

	for _, w := range workflows {
		if err := s.ExecuteWorkflow(ctx, w.ID); err != nil {
			log.Printf("scheduled workflow %q execution failed: %v", w.ID, err)
			// Continue processing remaining workflows even if one fails
		}
	}

	return nil
}

func (s *workflowService) GetWorkflowStatus(ctx context.Context, workflowID string) (*domain.Workflow, error) {
	return s.workflowRepo.FindByID(ctx, workflowID)
}
