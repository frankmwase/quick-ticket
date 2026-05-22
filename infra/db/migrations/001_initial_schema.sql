-- Phase 2 SQL Migration Script
CREATE TYPE ticket_status AS ENUM ('PENDING', 'ISSUED', 'VALIDATED', 'REVOKED');
CREATE TYPE template_type AS ENUM ('JSON_LAYOUT', 'PDF_CUSTOM');

CREATE TABLE tenants (
    id UUID PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    api_key_hash VARCHAR(64) UNIQUE NOT NULL,
    settings JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE ticket_batches (
    id UUID PRIMARY KEY,
    tenant_id UUID REFERENCES tenants(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    scheduled_release TIMESTAMP WITH TIME ZONE,
    auto_send_email BOOLEAN DEFAULT FALSE,
    geofence_config JSONB, -- Fields: latitude, longitude, radius_meters
    render_spec JSONB,     -- Fields: template_type, qr_code_placement (x, y, size)
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE tickets (
    id UUID PRIMARY KEY,
    tenant_id UUID REFERENCES tenants(id) ON DELETE CASCADE,
    batch_id UUID REFERENCES ticket_batches(id) ON DELETE SET NULL,
    owner_id VARCHAR(255) NOT NULL,
    managed_by VARCHAR(255), -- ID of delegator managing for someone else
    secure_token VARCHAR(255) UNIQUE NOT NULL, -- Cryptographically signed or randomized
    status ticket_status NOT NULL DEFAULT 'PENDING',
    history JSONB NOT NULL DEFAULT '[]', -- Audit trail tracks state mutations
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE idempotency_keys (
    key VARCHAR(255) PRIMARY KEY,
    response_code INT NOT NULL,
    response_body TEXT NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL
);
