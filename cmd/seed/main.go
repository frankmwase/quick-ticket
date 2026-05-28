package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"time"

	"quick-ticket/domain"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://yieldyeti:yieldyeti@localhost:5432/quickticket?sslmode=disable"
	}

	ctx := context.Background()
	db, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer db.Close()

	// demo tenant
	tenantID := "11111111-1111-1111-1111-111111111111"
	apiKey := "demo-api-key-123"
	h := sha256.Sum256([]byte(apiKey))
	apiKeyHash := hex.EncodeToString(h[:])

	tenant := &domain.Tenant{
		ID:         tenantID,
		Name:       "Demo Corp",
		APIKeyHash: apiKeyHash,
		CreatedAt:  time.Now().UTC(),
	}

	
	err = db.QueryRow(ctx, `
		INSERT INTO tenants (id, name, api_key_hash, created_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (api_key_hash) DO UPDATE SET name = EXCLUDED.name
		RETURNING id
	`, tenant.ID, tenant.Name, tenant.APIKeyHash, tenant.CreatedAt).Scan(&tenant.ID)

	if err != nil {
		log.Fatalf("Error inserting tenant: %v", err)
	}

	// demo user
	profileID := "22222222-2222-2222-2222-222222222222"
	_, err = db.Exec(ctx, `
		INSERT INTO user_profiles (id, tenant_id, email, full_name, role, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (tenant_id, email) DO NOTHING
	`, profileID, tenant.ID, "admin@democorp.com", "Demo Admin", "admin", time.Now().UTC())

	if err != nil {
		log.Fatalf("Error inserting user profile: %v", err)
	}

	fmt.Println("Seed completed successfully!")
	fmt.Printf("Tenant ID: %s\n", tenant.ID)
	fmt.Printf("API Key: %s\n", apiKey)
	fmt.Println("You can use these credentials to log in to the Web UI or TUI.")
}
