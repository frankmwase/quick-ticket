package domain

import (
	"time"
)

// Tenant represents a multi-tenancy aggregate root. Each tenant is an isolated
// organisation or API consumer with their own tickets, batches, and settings.
type Tenant struct {
	ID         string
	Name       string
	APIKeyHash string
	Settings   TenantSettings
	CreatedAt  time.Time
}

// TenantSettings stores per-tenant configuration as a flexible map.
// This mirrors the JSONB settings column and allows tenants to configure:
//   - webhook URLs, branding, default render specs, notification preferences, etc.
type TenantSettings struct {
	WebhookURL          string            `json:"webhook_url,omitempty"`
	WebhookSecret       string            `json:"webhook_secret,omitempty"`
	DefaultRenderSpec   []byte            `json:"default_render_spec,omitempty"`
	BrandingName        string            `json:"branding_name,omitempty"`
	BrandingLogoURL     string            `json:"branding_logo_url,omitempty"`
	NotificationDefaults NotificationConfig `json:"notification_defaults,omitempty"`
	MaxBatchSize        int               `json:"max_batch_size,omitempty"`
	EnableGeofencing    bool              `json:"enable_geofencing,omitempty"`
	CustomFields        map[string]string `json:"custom_fields,omitempty"`
}

// NotificationConfig specifies default notification behaviour for a tenant.
type NotificationConfig struct {
	EmailEnabled bool   `json:"email_enabled"`
	SMSEnabled   bool   `json:"sms_enabled"`
	FromEmail    string `json:"from_email,omitempty"`
	FromName     string `json:"from_name,omitempty"`
}

// TenantRepository defines the outbound port for tenant persistence.
type TenantRepository interface {
	FindByID(id string) (*Tenant, error)
	FindByAPIKeyHash(hash string) (*Tenant, error)
	Save(tenant *Tenant) error
	Update(tenant *Tenant) error
}
