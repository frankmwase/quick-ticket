package services

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"time"

	"quick-ticket/domain"

	"github.com/google/uuid"
)

type tenantService struct {
	tenantRepo domain.TenantRepository
}

// NewTenantService constructs a tenant management service.
func NewTenantService(tr domain.TenantRepository) *tenantService {
	return &tenantService{tenantRepo: tr}
}

func (s *tenantService) CreateTenant(ctx context.Context, tenant *domain.Tenant) (apiKey string, err error) {
	if tenant.Name == "" {
		return "", &domain.ErrValidation{Field: "name", Message: "tenant name is required"}
	}

	tenant.ID = uuid.New().String()
	tenant.CreatedAt = time.Now().UTC()

	// Generate a raw API key and store only the hash
	rawKey, err := generateAPIKey()
	if err != nil {
		return "", err
	}

	tenant.APIKeyHash = hashKey(rawKey)

	// Apply defaults
	if tenant.Settings.MaxBatchSize == 0 {
		tenant.Settings.MaxBatchSize = 10000
	}

	if err := s.tenantRepo.Save(tenant); err != nil {
		return "", err
	}

	// Return the raw key only once — it's never stored in plaintext
	return rawKey, nil
}

func (s *tenantService) GetTenant(ctx context.Context, id string) (*domain.Tenant, error) {
	return s.tenantRepo.FindByID(id)
}

func (s *tenantService) UpdateSettings(ctx context.Context, id string, settings domain.TenantSettings) error {
	tenant, err := s.tenantRepo.FindByID(id)
	if err != nil {
		return err
	}

	tenant.Settings = settings
	return s.tenantRepo.Update(tenant)
}

func (s *tenantService) RotateAPIKey(ctx context.Context, id string) (string, error) {
	tenant, err := s.tenantRepo.FindByID(id)
	if err != nil {
		return "", err
	}

	rawKey, err := generateAPIKey()
	if err != nil {
		return "", err
	}

	tenant.APIKeyHash = hashKey(rawKey)
	if err := s.tenantRepo.Update(tenant); err != nil {
		return "", err
	}

	return rawKey, nil
}

func generateAPIKey() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return "qt_" + hex.EncodeToString(bytes), nil
}

func hashKey(key string) string {
	h := sha256.Sum256([]byte(key))
	return hex.EncodeToString(h[:])
}
