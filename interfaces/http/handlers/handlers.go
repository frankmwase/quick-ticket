package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"quick-ticket/app/services"
	"quick-ticket/domain"
	"quick-ticket/interfaces/http/middleware"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// ──────────────────────────────────────────────────────────────────────────────
// Ticket Handler — Core ticket operations
// ──────────────────────────────────────────────────────────────────────────────

type TicketHandler struct {
	ticketService domain.TicketService
	batchRepo     domain.BatchRepository
}

func NewTicketHandler(ts domain.TicketService, br domain.BatchRepository) *TicketHandler {
	return &TicketHandler{ticketService: ts, batchRepo: br}
}

type GenerateRequest struct {
	BatchID    string `json:"batch_id"`
	Count      int    `json:"count"`
	OwnerID    string `json:"owner_id"`
	ManagedBy  string `json:"managed_by,omitempty"`
	AutoEmail  bool   `json:"auto_email"`
	TargetMail string `json:"target_mail,omitempty"`
}

type VerifyRequest struct {
	Token     string  `json:"token"`
	Latitude  float64 `json:"latitude,omitempty"`
	Longitude float64 `json:"longitude,omitempty"`
}

type RevokeRequest struct {
	TicketID string `json:"ticket_id"`
	ActorID  string `json:"actor_id"`
	Reason   string `json:"reason"`
}

type StatusUpdateRequest struct {
	Status  domain.TicketStatus `json:"status"`
	ActorID string              `json:"actor_id"`
	Reason  string              `json:"reason"`
}

func (h *TicketHandler) GenerateTickets(w http.ResponseWriter, r *http.Request) {
	var req GenerateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}

	if req.Count <= 0 || req.Count > 10000 {
		writeError(w, http.StatusBadRequest, "count must be between 1 and 10000")
		return
	}

	tenant := middleware.TenantFromContext(r.Context())
	tenantID := ""
	if tenant != nil {
		tenantID = tenant.ID
	}

	// Auto-fill owner_id from tenant if not provided
	ownerID := req.OwnerID
	if ownerID == "" {
		ownerID = tenantID
	}

	// Auto-create a batch if none specified
	batchID := req.BatchID
	if batchID == "" {
		newBatch := &domain.TicketBatch{
			ID:       uuid.New().String(),
			TenantID: tenantID,
			Title:    "Auto-generated batch",
		}
		if err := h.batchRepo.Save(r.Context(), newBatch); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to create batch: "+err.Error())
			return
		}
		batchID = newBatch.ID
	}

	bulkReq := domain.BulkGenerateRequest{
		TenantID:   tenantID,
		BatchID:    batchID,
		Count:      req.Count,
		OwnerID:    ownerID,
		ManagedBy:  req.ManagedBy,
		AutoEmail:  req.AutoEmail,
		TargetMail: req.TargetMail,
	}

	tickets, err := h.ticketService.GenerateBulk(r.Context(), bulkReq)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"count":   len(tickets),
		"tickets": tickets,
	})
}

func (h *TicketHandler) VerifyTicket(w http.ResponseWriter, r *http.Request) {
	var req VerifyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}

	if req.Token == "" {
		writeError(w, http.StatusBadRequest, "token is required")
		return
	}

	result, err := h.ticketService.VerifyTicket(r.Context(), domain.VerificationContext{
		Token:     req.Token,
		Latitude:  req.Latitude,
		Longitude: req.Longitude,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, result)
}

func (h *TicketHandler) RevokeTicket(w http.ResponseWriter, r *http.Request) {
	var req RevokeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}

	if req.TicketID == "" {
		writeError(w, http.StatusBadRequest, "ticket_id is required")
		return
	}

	err := h.ticketService.UpdateStatus(r.Context(), req.TicketID, domain.StatusRevoked, req.ActorID, req.Reason)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"status":  "revoked",
		"ticket_id": req.TicketID,
	})
}

func (h *TicketHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "ticket id is required")
		return
	}

	var req StatusUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}

	err := h.ticketService.UpdateStatus(r.Context(), id, req.Status, req.ActorID, req.Reason)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"status":    string(req.Status),
		"ticket_id": id,
	})
}

// ──────────────────────────────────────────────────────────────────────────────
// Batch Handler — Batch lifecycle operations
// ──────────────────────────────────────────────────────────────────────────────

type BatchHandler struct {
	batchRepo domain.BatchRepository
}

func NewBatchHandler(br domain.BatchRepository) *BatchHandler {
	return &BatchHandler{batchRepo: br}
}

type CreateBatchRequest struct {
	Title            string          `json:"title"`
	ScheduledRelease *time.Time      `json:"scheduled_release,omitempty"`
	AutoSendEmail    bool            `json:"auto_send_email"`
	GeofenceConfig   *domain.Geofence `json:"geofence_config,omitempty"`
	RenderSpec       json.RawMessage `json:"render_spec,omitempty"`
}

func (h *BatchHandler) CreateBatch(w http.ResponseWriter, r *http.Request) {
	var req CreateBatchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}

	if req.Title == "" {
		writeError(w, http.StatusBadRequest, "title is required")
		return
	}

	tenant := middleware.TenantFromContext(r.Context())
	tenantID := ""
	if tenant != nil {
		tenantID = tenant.ID
	}

	batch := &domain.TicketBatch{
		ID:               uuid.New().String(),
		TenantID:         tenantID,
		Title:            req.Title,
		ScheduledRelease: req.ScheduledRelease,
		AutoSendEmail:    req.AutoSendEmail,
		Geofence:         req.GeofenceConfig,
		RenderSpec:       req.RenderSpec,
		CreatedAt:        time.Now().UTC(),
	}

	if err := h.batchRepo.Save(r.Context(), batch); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, batch)
}

func (h *BatchHandler) GetBatchStatus(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "batch id is required")
		return
	}

	batch, err := h.batchRepo.FindByID(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, batch)
}

// ──────────────────────────────────────────────────────────────────────────────
// Workflow Handler — Workflow/scheduling operations
// ──────────────────────────────────────────────────────────────────────────────

type WorkflowHandler struct {
	workflowService domain.WorkflowService
}

func NewWorkflowHandler(ws domain.WorkflowService) *WorkflowHandler {
	return &WorkflowHandler{workflowService: ws}
}

type CreateWorkflowRequest struct {
	BatchID     string              `json:"batch_id"`
	Name        string              `json:"name"`
	ScheduledAt *time.Time          `json:"scheduled_at,omitempty"`
	Steps       []domain.WorkflowStep `json:"steps"`
	MaxRetries  int                 `json:"max_retries"`
}

func (h *WorkflowHandler) CreateWorkflow(w http.ResponseWriter, r *http.Request) {
	var req CreateWorkflowRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}

	tenant := middleware.TenantFromContext(r.Context())
	tenantID := ""
	if tenant != nil {
		tenantID = tenant.ID
	}

	workflow := &domain.Workflow{
		ID:          uuid.New().String(),
		TenantID:    tenantID,
		BatchID:     req.BatchID,
		Name:        req.Name,
		ScheduledAt: req.ScheduledAt,
		Steps:       req.Steps,
		MaxRetries:  req.MaxRetries,
	}

	if err := h.workflowService.CreateWorkflow(r.Context(), workflow); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, workflow)
}

func (h *WorkflowHandler) ExecuteWorkflow(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "workflow id is required")
		return
	}

	if err := h.workflowService.ExecuteWorkflow(r.Context(), id); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "executed", "workflow_id": id})
}

func (h *WorkflowHandler) CancelWorkflow(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "workflow id is required")
		return
	}

	if err := h.workflowService.CancelWorkflow(r.Context(), id); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "cancelled", "workflow_id": id})
}

func (h *WorkflowHandler) GetWorkflowStatus(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "workflow id is required")
		return
	}

	workflow, err := h.workflowService.GetWorkflowStatus(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, workflow)
}

// ──────────────────────────────────────────────────────────────────────────────
// Extension Handler — MW-JSON checkout
// ──────────────────────────────────────────────────────────────────────────────

type ExtensionHandler struct {
	orchestrator *services.MonetizedTicketOrchestrator
}

func NewExtensionHandler(orchestrator *services.MonetizedTicketOrchestrator) *ExtensionHandler {
	return &ExtensionHandler{orchestrator: orchestrator}
}

type CheckoutRequest struct {
	Transaction domain.MWJsonTransaction `json:"transaction"`
	TicketReq   GenerateRequest          `json:"ticket_request"`
}

func (h *ExtensionHandler) Checkout(w http.ResponseWriter, r *http.Request) {
	var req CheckoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}

	tenant := middleware.TenantFromContext(r.Context())
	tenantID := ""
	if tenant != nil {
		tenantID = tenant.ID
	}

	bulkReq := domain.BulkGenerateRequest{
		TenantID:   tenantID,
		BatchID:    req.TicketReq.BatchID,
		Count:      req.TicketReq.Count,
		OwnerID:    req.TicketReq.OwnerID,
		ManagedBy:  req.TicketReq.ManagedBy,
		AutoEmail:  req.TicketReq.AutoEmail,
		TargetMail: req.TicketReq.TargetMail,
	}

	tickets, err := h.orchestrator.PurchaseTicketWithMWJson(r.Context(), &req.Transaction, bulkReq)
	if err != nil {
		writeError(w, http.StatusPaymentRequired, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"count":   len(tickets),
		"tickets": tickets,
	})
}

// ──────────────────────────────────────────────────────────────────────────────
// Health Handler — System health check
// ──────────────────────────────────────────────────────────────────────────────

type HealthHandler struct{}

func NewHealthHandler() *HealthHandler { return &HealthHandler{} }

func (h *HealthHandler) Health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":  "healthy",
		"service": "quick-ticket-baas",
		"version": "1.0.0",
		"time":    time.Now().UTC(),
	})
}

// ──────────────────────────────────────────────────────────────────────────────
// Helpers
// ──────────────────────────────────────────────────────────────────────────────

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
