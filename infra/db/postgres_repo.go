package db

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"quick-ticket/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ──────────────────────────────────────────────────────────────────────────────
// PostgresTicketRepo — High-performance ticket repository using pgx.
// Uses CopyFrom for bulk inserts to bypass individual INSERT overhead.
// ──────────────────────────────────────────────────────────────────────────────

type PostgresTicketRepo struct {
	pool *pgxpool.Pool
}

func NewPostgresTicketRepo(pool *pgxpool.Pool) domain.TicketRepository {
	return &PostgresTicketRepo{pool: pool}
}

func (r *PostgresTicketRepo) SaveBulk(ctx context.Context, tickets []*domain.Ticket) error {
	columns := []string{"id", "tenant_id", "batch_id", "owner_id", "managed_by", "secure_token", "status", "history", "created_at", "updated_at"}

	rows := make([][]interface{}, len(tickets))
	now := time.Now().UTC()

	for i, t := range tickets {
		historyJSON, _ := json.Marshal(t.History)
		rows[i] = []interface{}{
			t.ID, t.TenantID, t.BatchID, t.OwnerID, t.ManagedBy,
			t.SecureToken, string(t.Status), historyJSON, t.CreatedAt, now,
		}
	}

	copyCount, err := r.pool.CopyFrom(
		ctx,
		pgx.Identifier{"tickets"},
		columns,
		pgx.CopyFromRows(rows),
	)

	if err != nil {
		return fmt.Errorf("pgx CopyFrom failed: %w", err)
	}

	if int(copyCount) != len(tickets) {
		return fmt.Errorf("expected %d rows inserted, got %d", len(tickets), copyCount)
	}

	return nil
}

func (r *PostgresTicketRepo) FindByID(ctx context.Context, id string) (*domain.Ticket, error) {
	query := `SELECT id, tenant_id, batch_id, owner_id, managed_by, secure_token, status, history, created_at
	           FROM tickets WHERE id = $1`

	ticket := &domain.Ticket{}
	var historyJSON []byte
	var status string

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&ticket.ID, &ticket.TenantID, &ticket.BatchID, &ticket.OwnerID,
		&ticket.ManagedBy, &ticket.SecureToken, &status, &historyJSON, &ticket.CreatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, &domain.ErrNotFound{Entity: "ticket", ID: id}
		}
		return nil, err
	}

	ticket.Status = domain.TicketStatus(status)
	_ = json.Unmarshal(historyJSON, &ticket.History)

	return ticket, nil
}

func (r *PostgresTicketRepo) FindByToken(ctx context.Context, token string) (*domain.Ticket, error) {
	query := `SELECT id, tenant_id, batch_id, owner_id, managed_by, secure_token, status, history, created_at
	           FROM tickets WHERE secure_token = $1`

	ticket := &domain.Ticket{}
	var historyJSON []byte
	var status string

	err := r.pool.QueryRow(ctx, query, token).Scan(
		&ticket.ID, &ticket.TenantID, &ticket.BatchID, &ticket.OwnerID,
		&ticket.ManagedBy, &ticket.SecureToken, &status, &historyJSON, &ticket.CreatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, &domain.ErrNotFound{Entity: "ticket", ID: token}
		}
		return nil, err
	}

	ticket.Status = domain.TicketStatus(status)
	_ = json.Unmarshal(historyJSON, &ticket.History)

	return ticket, nil
}

func (r *PostgresTicketRepo) Update(ctx context.Context, ticket *domain.Ticket) error {
	historyJSON, _ := json.Marshal(ticket.History)

	query := `UPDATE tickets SET status = $1, history = $2, managed_by = $3, updated_at = $4 WHERE id = $5`
	_, err := r.pool.Exec(ctx, query, string(ticket.Status), historyJSON, ticket.ManagedBy, time.Now().UTC(), ticket.ID)
	return err
}

// ──────────────────────────────────────────────────────────────────────────────
// PostgresBatchRepo — Batch repository with JSONB geofence/render spec support.
// ──────────────────────────────────────────────────────────────────────────────

type PostgresBatchRepo struct {
	pool *pgxpool.Pool
}

func NewPostgresBatchRepo(pool *pgxpool.Pool) domain.BatchRepository {
	return &PostgresBatchRepo{pool: pool}
}

func (r *PostgresBatchRepo) FindByID(ctx context.Context, id string) (*domain.TicketBatch, error) {
	query := `SELECT id, tenant_id, title, scheduled_release, auto_send_email, geofence_config, render_spec, created_at
	           FROM ticket_batches WHERE id = $1`

	batch := &domain.TicketBatch{}
	var geofenceJSON, renderSpecJSON []byte

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&batch.ID, &batch.TenantID, &batch.Title, &batch.ScheduledRelease,
		&batch.AutoSendEmail, &geofenceJSON, &renderSpecJSON, &batch.CreatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, &domain.ErrNotFound{Entity: "batch", ID: id}
		}
		return nil, err
	}

	if geofenceJSON != nil {
		var geo domain.Geofence
		if err := json.Unmarshal(geofenceJSON, &geo); err == nil {
			batch.Geofence = &geo
		}
	}

	batch.RenderSpec = renderSpecJSON
	return batch, nil
}

func (r *PostgresBatchRepo) Save(ctx context.Context, batch *domain.TicketBatch) error {
	var geofenceJSON []byte
	if batch.Geofence != nil {
		geofenceJSON, _ = json.Marshal(batch.Geofence)
	}

	query := `INSERT INTO ticket_batches (id, tenant_id, title, scheduled_release, auto_send_email, geofence_config, render_spec, created_at)
	           VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err := r.pool.Exec(ctx, query,
		batch.ID, batch.TenantID, batch.Title, batch.ScheduledRelease,
		batch.AutoSendEmail, geofenceJSON, batch.RenderSpec, time.Now().UTC(),
	)
	return err
}

// ──────────────────────────────────────────────────────────────────────────────
// PostgresWorkflowRepo — Workflow persistence with step JSONB storage.
// ──────────────────────────────────────────────────────────────────────────────

type PostgresWorkflowRepo struct {
	pool *pgxpool.Pool
}

func NewPostgresWorkflowRepo(pool *pgxpool.Pool) domain.WorkflowRepository {
	return &PostgresWorkflowRepo{pool: pool}
}

func (r *PostgresWorkflowRepo) Save(ctx context.Context, workflow *domain.Workflow) error {
	stepsJSON, _ := json.Marshal(workflow.Steps)

	query := `INSERT INTO workflows (id, tenant_id, batch_id, name, status, steps, scheduled_at, created_at, retry_count, max_retries)
	           VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`

	_, err := r.pool.Exec(ctx, query,
		workflow.ID, workflow.TenantID, workflow.BatchID, workflow.Name,
		string(workflow.Status), stepsJSON, workflow.ScheduledAt,
		workflow.CreatedAt, workflow.RetryCount, workflow.MaxRetries,
	)
	return err
}

func (r *PostgresWorkflowRepo) FindByID(ctx context.Context, id string) (*domain.Workflow, error) {
	query := `SELECT id, tenant_id, batch_id, name, status, steps, scheduled_at, executed_at, completed_at, created_at, retry_count, max_retries
	           FROM workflows WHERE id = $1`

	w := &domain.Workflow{}
	var stepsJSON []byte
	var status string

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&w.ID, &w.TenantID, &w.BatchID, &w.Name, &status,
		&stepsJSON, &w.ScheduledAt, &w.ExecutedAt, &w.CompletedAt,
		&w.CreatedAt, &w.RetryCount, &w.MaxRetries,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, &domain.ErrNotFound{Entity: "workflow", ID: id}
		}
		return nil, err
	}

	w.Status = domain.WorkflowStatus(status)
	_ = json.Unmarshal(stepsJSON, &w.Steps)

	return w, nil
}

func (r *PostgresWorkflowRepo) FindPendingByTenant(ctx context.Context, tenantID string) ([]*domain.Workflow, error) {
	query := `SELECT id, tenant_id, batch_id, name, status, steps, scheduled_at, created_at, retry_count, max_retries
	           FROM workflows WHERE tenant_id = $1 AND status = 'PENDING' ORDER BY created_at ASC`

	rows, err := r.pool.Query(ctx, query, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var workflows []*domain.Workflow
	for rows.Next() {
		w := &domain.Workflow{}
		var stepsJSON []byte
		var status string

		if err := rows.Scan(&w.ID, &w.TenantID, &w.BatchID, &w.Name, &status,
			&stepsJSON, &w.ScheduledAt, &w.CreatedAt, &w.RetryCount, &w.MaxRetries); err != nil {
			return nil, err
		}
		w.Status = domain.WorkflowStatus(status)
		_ = json.Unmarshal(stepsJSON, &w.Steps)
		workflows = append(workflows, w)
	}

	return workflows, nil
}

func (r *PostgresWorkflowRepo) FindDueForExecution(ctx context.Context) ([]*domain.Workflow, error) {
	query := `SELECT id, tenant_id, batch_id, name, status, steps, scheduled_at, created_at, retry_count, max_retries
	           FROM workflows WHERE status = 'PENDING' AND (scheduled_at IS NULL OR scheduled_at <= NOW())
	           ORDER BY scheduled_at ASC NULLS FIRST`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var workflows []*domain.Workflow
	for rows.Next() {
		w := &domain.Workflow{}
		var stepsJSON []byte
		var status string

		if err := rows.Scan(&w.ID, &w.TenantID, &w.BatchID, &w.Name, &status,
			&stepsJSON, &w.ScheduledAt, &w.CreatedAt, &w.RetryCount, &w.MaxRetries); err != nil {
			return nil, err
		}
		w.Status = domain.WorkflowStatus(status)
		_ = json.Unmarshal(stepsJSON, &w.Steps)
		workflows = append(workflows, w)
	}

	return workflows, nil
}

func (r *PostgresWorkflowRepo) Update(ctx context.Context, workflow *domain.Workflow) error {
	stepsJSON, _ := json.Marshal(workflow.Steps)

	query := `UPDATE workflows SET status = $1, steps = $2, executed_at = $3, completed_at = $4, retry_count = $5
	           WHERE id = $6`

	_, err := r.pool.Exec(ctx, query,
		string(workflow.Status), stepsJSON, workflow.ExecutedAt,
		workflow.CompletedAt, workflow.RetryCount, workflow.ID,
	)
	return err
}

// ──────────────────────────────────────────────────────────────────────────────
// PostgresTenantRepo — Tenant persistence with JSONB settings.
// ──────────────────────────────────────────────────────────────────────────────

type PostgresTenantRepo struct {
	pool *pgxpool.Pool
}

func NewPostgresTenantRepo(pool *pgxpool.Pool) domain.TenantRepository {
	return &PostgresTenantRepo{pool: pool}
}

func (r *PostgresTenantRepo) FindByID(id string) (*domain.Tenant, error) {
	query := `SELECT id, name, api_key_hash, settings, created_at FROM tenants WHERE id = $1`

	tenant := &domain.Tenant{}
	var settingsJSON []byte

	err := r.pool.QueryRow(context.Background(), query, id).Scan(
		&tenant.ID, &tenant.Name, &tenant.APIKeyHash, &settingsJSON, &tenant.CreatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, &domain.ErrNotFound{Entity: "tenant", ID: id}
		}
		return nil, err
	}

	_ = json.Unmarshal(settingsJSON, &tenant.Settings)
	return tenant, nil
}

func (r *PostgresTenantRepo) FindByAPIKeyHash(hash string) (*domain.Tenant, error) {
	query := `SELECT id, name, api_key_hash, settings, created_at FROM tenants WHERE api_key_hash = $1`

	tenant := &domain.Tenant{}
	var settingsJSON []byte

	err := r.pool.QueryRow(context.Background(), query, hash).Scan(
		&tenant.ID, &tenant.Name, &tenant.APIKeyHash, &settingsJSON, &tenant.CreatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, &domain.ErrNotFound{Entity: "tenant", ID: hash}
		}
		return nil, err
	}

	_ = json.Unmarshal(settingsJSON, &tenant.Settings)
	return tenant, nil
}

func (r *PostgresTenantRepo) Save(tenant *domain.Tenant) error {
	settingsJSON, _ := json.Marshal(tenant.Settings)

	query := `INSERT INTO tenants (id, name, api_key_hash, settings, created_at)
	           VALUES ($1, $2, $3, $4, $5)`

	_, err := r.pool.Exec(context.Background(), query,
		tenant.ID, tenant.Name, tenant.APIKeyHash, settingsJSON, time.Now().UTC(),
	)
	return err
}

func (r *PostgresTenantRepo) Update(tenant *domain.Tenant) error {
	settingsJSON, _ := json.Marshal(tenant.Settings)

	query := `UPDATE tenants SET name = $1, settings = $2 WHERE id = $3`
	_, err := r.pool.Exec(context.Background(), query, tenant.Name, settingsJSON, tenant.ID)
	return err
}
