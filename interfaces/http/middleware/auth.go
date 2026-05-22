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

// Auth is a middleware that validates the API key from the Authorization header
// and injects the resolved tenant into the request context.
func Auth(tenantRepo domain.TenantRepository) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, `{"error":"missing Authorization header"}`, http.StatusUnauthorized)
				return
			}

			// Expect "Bearer <api-key>"
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				http.Error(w, `{"error":"invalid Authorization format, expected Bearer <key>"}`, http.StatusUnauthorized)
				return
			}

			apiKey := parts[1]
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
