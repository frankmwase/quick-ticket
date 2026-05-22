# Quick Ticket —  Backend as a Service

A  headless ticketing BaaS built in Go using strict Domain Driven Design. Like MedusaJS but for tickets. third parties build any UI on top of the API.

## Features

### Core Ticketing
- **High-speed bulk generation** — produce thousands of tickets in a single atomic call via `pgx.CopyFrom`
- **Cryptographically secure tokens** — every ticket gets a 256-bit random token for verification
- **Full lifecycle management** — `PENDING → ISSUED → VALIDATED → REVOKED` state machine with audit trails
- **Geofencing** — Haversine-formula location validation before ticket acceptance
- **Delegation** — manage tickets on behalf of other individuals via `managed_by`

### Distribution & Rendering
- **PDF generation** — custom PDF tickets with configurable QR code placement via `go-pdf/fpdf`
- **QR codes** — auto-generated and embedded in PDFs using the secure token
- **Email delivery** — SMTP with MIME multipart PDF attachments
- **Print layouts** — JSON POS or thermal printer specs returned in API responses
- **Render specs** — fully configurable per batch template with meta fields

### Workflows & Scheduling
- **Custom workflow pipelines** — chain steps: generate → render → email → webhook
- **Scheduled releases** — set future release times for ticket batches
- **Background scheduler** — worker processes due workflows every 30 seconds
- **Step-level error handling** — individual step failures with retry support

### Monetisation
- **Malawi Pay Standard (MW-JSON)** — native integration for Airtel Money, TNM Mpamba, bank transfers
- **Payment gated issuance** — tickets only generate after payment verification
- **Structural validation** — enforces MW-JSON spec (msgId, currency, participant IDs, TTL)
- **Idempotency keys** — prevents duplicate billing on flaky networks

### Multi-Tenancy & Security
- **Tenant isolation** — every record scoped by `tenant_id`
- **API key auth** — SHA-256 hashed keys, rotation support, raw key returned only at creation
- **Idempotency engine** — middleware intercepts duplicate POST requests and replays cached responses
- **CORS** — permissive headers for headless consumption
- **Webhook signatures** — HMAC-SHA256 signed payloads with exponential backoff retries

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                     interfaces/http                         │
│  handlers  │  middleware (auth, tenant, idempotency, CORS)  │
│  router.go                                                  │
└──────────────────────────┬──────────────────────────────────┘
                           │ depends on
┌──────────────────────────▼──────────────────────────────────┐
│                      app/services                           │
│  ticket_service  │  workflow_service  │  payment_orchestrator│
│  tenant_service                                             │
└──────────────────────────┬──────────────────────────────────┘
                           │ depends on
┌──────────────────────────▼──────────────────────────────────┐
│                        domain                               │
│  ticket  │  tenant  │  workflow  │  payment  │  errors      │
│  ports (TicketRepository, BatchRepository, PDFRenderer...)  │
└──────────────────────────┬──────────────────────────────────┘
                           │ implemented by
┌──────────────────────────▼──────────────────────────────────┐
│                         infra                               │
│  db/postgres_repo  │  delivery/smtp  │  delivery/webhook    │
│  payments/mwjson   │  rendering/pdf  │  security/idempotency│
└─────────────────────────────────────────────────────────────┘
```

All dependencies point **inward**. Infrastructure implements domain interfaces , no circular dependancy.

## Directory Structure

```
├── cmd/baas-core/
│   ├── main.go                          # Entrypoint with full DI wiring
│   └── main_bench_test.go               # Concurrent idempotency benchmark
├── domain/
│   ├── ticket.go                        # Ticket & Batch aggregate roots
│   ├── tenant.go                        # Multi-tenant aggregate root
│   ├── workflow.go                      # Workflow pipeline & scheduling
│   ├── payment.go                       # MW-JSON transaction models
│   ├── ports.go                         # Service & repository interfaces
│   └── errors.go                        # Strongly typed domain errors
├── app/
│   ├── services/
│   │   ├── ticket_service.go            # Bulk gen, verify, distribute, webhook dispatch
│   │   ├── workflow_service.go          # Pipeline executor & scheduler
│   │   ├── payment_orchestrator.go      # MW-JSON → ticket generation bridge
│   │   ├── tenant_service.go            # Tenant CRUD & API key management
│   │   └── ticket_service_test.go       # Unit tests with gomock
│   └── ports/
│       ├── repos.go                     # Inbound use-case interfaces
│       └── infrastructure.go            # Outbound infra contracts
├── infra/
│   ├── db/
│   │   ├── postgres_repo.go             # pgx CopyFrom bulk repo and all repos
│   │   └── migrations/
│   │       ├── 001_initial_schema.sql   # Core tables
│   │       └── 002_workflows_and_webhooks.sql
│   ├── delivery/
│   │   ├── smtp.go                      # SMTP + MIME PDF attachments
│   │   └── webhook.go                   # HMAC-signed webhook dispatcher
│   ├── payments/
│   │   ├── mwjson_provider.go           # MW-JSON gateway integration
│   │   └── mwjson_integration_test.go
│   ├── rendering/
│   │   └── pdf_renderer.go             # fpdf + QR code generator
│   └── security/
│       └── idempotency.go              # In-memory store with TTL cleanup
├── interfaces/http/
│   ├── handlers/handlers.go             # All REST handlers
│   ├── middleware/
│   │   ├── auth.go                      # Bearer token → tenant resolution
│   │   ├── tenant.go                    # Tenant scope guard and CORS
│   │   └── idempotency.go              # Duplicate request interception
│   └── router.go                        # chi router with all routes
├── Dockerfile                           # Multi-stage Alpine build
├── docker-compose.yml                   # Postgres + Redis + MailHog
└── .env.example
```

## API Reference

### Health
| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/health` | System health check (unauthenticated) |

### Batches
| Method | Endpoint | Description |
|--------|----------|-------------|
| `POST` | `/api/v1/batches` | Create batch with geofence/render config |
| `GET` | `/api/v1/batches/:id/status` | Get batch details and config |

### Tickets
| Method | Endpoint | Description |
|--------|----------|-------------|
| `POST` | `/api/v1/tickets/generate` | Bulk generate (supports `Idempotency-Key`) |
| `POST` | `/api/v1/tickets/verify` | Verify token + optional geofence check |
| `POST` | `/api/v1/tickets/revoke` | Revoke a ticket |
| `PATCH` | `/api/v1/tickets/:id/status` | Update status (VALIDATED, REVOKED, etc.) |

### Workflows
| Method | Endpoint | Description |
|--------|----------|-------------|
| `POST` | `/api/v1/workflows` | Create workflow pipeline |
| `POST` | `/api/v1/workflows/:id/execute` | Trigger workflow execution |
| `POST` | `/api/v1/workflows/:id/cancel` | Cancel a pending workflow |
| `GET` | `/api/v1/workflows/:id/status` | Get workflow + step statuses |

### Extensions
| Method | Endpoint | Description |
|--------|----------|-------------|
| `POST` | `/api/v1/extensions/checkout` | MW-JSON payment → ticket generation |

## Quick Start

### 1. Start Infrastructure

```bash
docker compose up -d postgres redis mailhog
```

### 2. Run Migrations

```bash
psql $DATABASE_URL < infra/db/migrations/001_initial_schema.sql
psql $DATABASE_URL < infra/db/migrations/002_workflows_and_webhooks.sql
```

### 3. Run the Server

```bash
cp .env.example .env
# Edit .env with your values
go run ./cmd/baas-core
```

### 4. Generate Tickets

```bash
# Create a batch
curl -X POST http://localhost:8080/api/v1/batches \
  -H "Authorization: Bearer qt_your-api-key" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Lake of Stars Festival - VIP Pass",
    "auto_send_email": false,
    "geofence_config": {
      "lat": -13.98,
      "lng": 33.78,
      "radius_meters": 500
    },
    "render_spec": {
      "template_type": "pdf_custom",
      "qr_code_config": {
        "page": 1,
        "coordinate_x": 140,
        "coordinate_y": 680,
        "size_pixels": 250
      },
      "meta_fields": {
        "event_name": "Lake of Stars Festival - VIP Pass",
        "gate_open": "18:00 CAT",
        "seat_row": "A4"
      }
    }
  }'

# Bulk generate 10 tickets
curl -X POST http://localhost:8080/api/v1/tickets/generate \
  -H "Authorization: Bearer qt_your-api-key" \
  -H "Content-Type: application/json" \
  -H "Idempotency-Key: unique-request-id-001" \
  -d '{
    "batch_id": "<batch-id>",
    "count": 10,
    "owner_id": "user-123"
  }'

# Verify a ticket with location
curl -X POST http://localhost:8080/api/v1/tickets/verify \
  -H "Authorization: Bearer qt_your-api-key" \
  -H "Content-Type: application/json" \
  -d '{
    "token": "<secure-token>",
    "latitude": -13.9801,
    "longitude": 33.7801
  }'
```

### 5. MW-JSON Payment Checkout

```bash
curl -X POST http://localhost:8080/api/v1/extensions/checkout \
  -H "Authorization: Bearer qt_your-api-key" \
  -H "Content-Type: application/json" \
  -d '{
    "transaction": {
      "mwVersion": "1.0",
      "header": {
        "msgId": "tx-001",
        "timestamp": "2026-05-22T14:00:00Z",
        "ttl": 300,
        "idempotencyKey": "pay-001"
      },
      "payload": {
        "amount": 5000,
        "currency": "MWK",
        "type": "P2M",
        "sender": { "id": "+265999123456", "idType": "MSISDN" },
        "receiver": { "id": "MERCHANT-001", "idType": "MERCHANT_CODE" }
      }
    },
    "ticket_request": {
      "batch_id": "<batch-id>",
      "count": 2,
      "owner_id": "user-456"
    }
  }'
```

## Webhook Events

When a ticket changes state, registered webhook URLs receive signed JSON callbacks:

```json
{
  "event": "ticket.VALIDATED",
  "tenant_id": "0df6e85c-19c2-40b9-8bc3-3b12368945ff",
  "timestamp": 1782348900,
  "data": {
    "ticket_id": "7ca34b22-1244-4f9e-a010-0987162534de",
    "from_status": "ISSUED",
    "to_status": "VALIDATED",
    "actor_id": "OP-901",
    "occurred_at": "2026-05-22T14:08:00Z"
  }
}
```

Events dispatched: `ticket.batch_generated`, `ticket.ISSUED`, `ticket.VALIDATED`, `ticket.REVOKED`, `workflow.step_completed`

Signatures are verified via the `X-Webhook-Signature` header (HMAC-SHA256).

## Testing

```bash
# Unit tests
go test ./... -v

# Benchmarks
go test -bench=. -benchmem ./cmd/baas-core/

# Results: ~860 ops with concurrent idempotency at 1.38ms/op
```

## Sample Ticket JSON

```json
{
  "ticket_id": "8fa53bb2-33a5-48fa-bb6b-a010d29388df",
  "secure_token": "a4f89d31e90b8fbc01de47ff6d2b3882",
  "status": "ISSUED",
  "render_spec": {
    "template_type": "pdf_custom",
    "qr_code_config": {
      "page": 1,
      "coordinate_x": 140,
      "coordinate_y": 680,
      "size_pixels": 250
    },
    "meta_fields": {
      "event_name": "Lake of Stars Festival - VIP Pass",
      "gate_open": "18:00 CAT",
      "seat_row": "A4"
    }
  }
}
```

## License

MIT
