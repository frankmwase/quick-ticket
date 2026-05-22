-- Phase 2b: Workflow tables
CREATE TYPE workflow_status AS ENUM ('PENDING', 'ACTIVE', 'COMPLETED', 'FAILED', 'CANCELLED');

CREATE TABLE workflows (
    id UUID PRIMARY KEY,
    tenant_id UUID REFERENCES tenants(id) ON DELETE CASCADE,
    batch_id UUID REFERENCES ticket_batches(id) ON DELETE SET NULL,
    name VARCHAR(255) NOT NULL,
    status workflow_status NOT NULL DEFAULT 'PENDING',
    steps JSONB NOT NULL DEFAULT '[]',
    scheduled_at TIMESTAMP WITH TIME ZONE,
    executed_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    retry_count INT DEFAULT 0,
    max_retries INT DEFAULT 3
);

CREATE INDEX idx_workflows_tenant_status ON workflows(tenant_id, status);
CREATE INDEX idx_workflows_scheduled ON workflows(status, scheduled_at) WHERE status = 'PENDING';

-- Webhook subscription registry
CREATE TABLE webhook_subscriptions (
    id UUID PRIMARY KEY,
    tenant_id UUID REFERENCES tenants(id) ON DELETE CASCADE,
    url TEXT NOT NULL,
    secret VARCHAR(255),
    events TEXT[] NOT NULL DEFAULT '{}', -- e.g. {'ticket.validated', 'ticket.revoked'}
    active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Audit log for all state transitions
CREATE TABLE audit_log (
    id BIGSERIAL PRIMARY KEY,
    tenant_id UUID REFERENCES tenants(id) ON DELETE CASCADE,
    entity_type VARCHAR(50) NOT NULL,  -- 'ticket', 'workflow', 'batch'
    entity_id UUID NOT NULL,
    action VARCHAR(50) NOT NULL,       -- 'status_change', 'created', 'deleted'
    actor_id VARCHAR(255),
    old_value JSONB,
    new_value JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_audit_log_entity ON audit_log(entity_type, entity_id);
CREATE INDEX idx_audit_log_tenant ON audit_log(tenant_id, created_at DESC);
