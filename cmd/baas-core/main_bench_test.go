package main_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"quick-ticket/interfaces/http/middleware"
)

func BenchmarkConcurrentIdempotency(b *testing.B) {
	r := chi.NewRouter()
	
	// A simple mock handler that simulates an expensive operation 
	// and returns 201 Created.
	counter := 0
	var mu sync.Mutex

	handler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		time.Sleep(5 * time.Millisecond) // Simulate DB processing
		mu.Lock()
		counter++
		mu.Unlock()
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"status":"success"}`))
	})

	r.With(middleware.Idempotency).Post("/generate", handler)

	server := httptest.NewServer(r)
	defer server.Close()

	// Pre-generate keys
	keys := make([]string, b.N)
	for i := 0; i < b.N; i++ {
		keys[i] = uuid.New().String()
	}

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		client := server.Client()
		i := 0
		for pb.Next() {
			if i >= len(keys) {
				i = 0
			}
			key := keys[i]
			i++

			req, _ := http.NewRequest(http.MethodPost, server.URL+"/generate", bytes.NewBufferString(`{}`))
			req.Header.Set("Idempotency-Key", key)

			resp, err := client.Do(req)
			if err != nil {
				b.Fatalf("request failed: %v", err)
			}
			resp.Body.Close()

			if resp.StatusCode != http.StatusCreated {
				b.Fatalf("expected 201 Created, got %d", resp.StatusCode)
			}
		}
	})
}
