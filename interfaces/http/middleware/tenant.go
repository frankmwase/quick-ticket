package middleware

import (
	"net/http"
)

// TenantScope enforces that every request within its scope has a resolved tenant.
// It must be applied AFTER the Auth middleware.
func TenantScope(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tenant := TenantFromContext(r.Context())
		if tenant == nil {
			http.Error(w, `{"error":"tenant context not found — ensure Auth middleware is applied"}`, http.StatusForbidden)
			return
		}

		// Inject tenant ID as a response header for debugging
		w.Header().Set("X-Tenant-ID", tenant.ID)

		next.ServeHTTP(w, r)
	})
}

// CORS adds permissive CORS headers for headless BaaS consumers.
func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Idempotency-Key, X-Tenant-ID, X-API-Key")
		w.Header().Set("Access-Control-Expose-Headers", "X-Tenant-ID, X-Request-ID")
		w.Header().Set("Access-Control-Max-Age", "86400")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
