package middleware

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strings"

	"quick-ticket/domain"
)

type contextKey string

const TenantContextKey contextKey = "tenant"

// TenantFromContext extracts the authenticated tenant from the request context.
func TenantFromContext(ctx context.Context) *domain.Tenant {
	t, _ := ctx.Value(TenantContextKey).(*domain.Tenant)
	return t
}


// It supports two authentication methods (tried in order):
//  1. Authorization: Bearer <api-key>       (standard)
//  2. X-API-Key: <api-key>                  (used by clients)
func Auth(tenantRepo domain.TenantRepository) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			apiKey := ""

			// Try Authorization: Bearer first
			if authHeader := r.Header.Get("Authorization"); authHeader != "" {
				parts := strings.SplitN(authHeader, " ", 2)
				if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
					apiKey = parts[1]
				}
			}

			// Fall back to X-API-Key header
			if apiKey == "" {
				apiKey = r.Header.Get("X-API-Key")
			}

			if apiKey == "" {
				http.Error(w, `{"error":"missing API key — provide Authorization: Bearer <key> or X-API-Key header"}`, http.StatusUnauthorized)
				return
			}

			keyHash := hashAPIKey(apiKey)

			tenant, err := tenantRepo.FindByAPIKeyHash(keyHash)
			if err != nil || tenant == nil {
				http.Error(w, `{"error":"invalid API key"}`, http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), TenantContextKey, tenant)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func hashAPIKey(key string) string {
	h := sha256.Sum256([]byte(key))
	return hex.EncodeToString(h[:])
}
