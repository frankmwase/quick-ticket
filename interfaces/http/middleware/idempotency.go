package middleware

import (
	"bytes"
	"net/http"
	"sync"
	"time"
)

type cacheEntry struct {
	statusCode int
	body       []byte
	expiresAt  time.Time
}

var (
	idempotencyCache = make(map[string]cacheEntry)
	cacheMutex       sync.RWMutex
)

type responseRecorder struct {
	http.ResponseWriter
	statusCode int
	body       *bytes.Buffer
}

func (rw *responseRecorder) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseRecorder) Write(b []byte) (int, error) {
	rw.body.Write(b)
	return rw.ResponseWriter.Write(b)
}

func Idempotency(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			next.ServeHTTP(w, r)
			return
		}

		idempotencyKey := r.Header.Get("Idempotency-Key")
		if idempotencyKey == "" {
			next.ServeHTTP(w, r)
			return
		}

		cacheMutex.RLock()
		entry, found := idempotencyCache[idempotencyKey]
		cacheMutex.RUnlock()

		if found && time.Now().Before(entry.expiresAt) {
			w.WriteHeader(entry.statusCode)
			w.Write(entry.body)
			return
		}

		recorder := &responseRecorder{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
			body:           &bytes.Buffer{},
		}

		next.ServeHTTP(recorder, r)

		cacheMutex.Lock()
		idempotencyCache[idempotencyKey] = cacheEntry{
			statusCode: recorder.statusCode,
			body:       recorder.body.Bytes(),
			expiresAt:  time.Now().Add(24 * time.Hour), // 24 hours TTL
		}
		cacheMutex.Unlock()
	})
}
