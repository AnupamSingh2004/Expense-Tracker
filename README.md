# Expense Tracker

A production-grade personal expense tracking application.

## Architecture

```
frontend (Next.js)  →  backend (Go/chi)  →  PostgreSQL
        ↑                    ↑
  TanStack Query        handler → service → repository
```

- **Backend:** Clean architecture — handlers are thin, all business logic lives in the service layer, repositories own all SQL.
- **Frontend:** App Router + TanStack Query for server-state caching and retry. Idempotency key generated per submission, rotated on success.
- **Money:** Stored as `int64` paise (1 INR = 100 paise) throughout — no floats anywhere.
- **Idempotency:** `Idempotency-Key` header + SHA-256 hash of request body persisted to `idempotency_keys` table. Replays return the cached response. Key reuse with a different body returns 409.

## Why PostgreSQL?

- ACID guarantees for financial data.
- Native JSONB for storing idempotency response payloads.
- `ON CONFLICT DO NOTHING` makes idempotency writes safe under concurrent retries.

## Why int64 for money?

Floating-point arithmetic is lossy. `0.1 + 0.2 ≠ 0.3` in IEEE 754. Paise as integers eliminate rounding errors entirely. The frontend converts user-entered rupees to paise before sending.

## Idempotency Strategy

1. Client generates a UUID v4 `Idempotency-Key` per form submission (stable for retries, rotated on success).
2. Backend computes `SHA-256(request body)` and checks `idempotency_keys` table.
3. **Hit:** Returns cached response — no DB write.
4. **Miss:** Creates expense, persists `(key, hash, response)` atomically.
5. **Conflict (same key, different body):** Returns 409 `IDEMPOTENCY_CONFLICT`.

## How to Run Locally

**Prerequisites:** Docker + Docker Compose

```bash
cp .env.example .env
docker compose up --build
```

- Frontend: http://localhost:3000
- Backend API: http://localhost:8080
- Health check: http://localhost:8080/health

## API Reference

### POST /expenses
Create an expense. Amount in **paise**.

```bash
curl -X POST http://localhost:8080/expenses \
  -H "Content-Type: application/json" \
  -H "Idempotency-Key: $(uuidgen)" \
  -d '{"amount": 5000, "category": "food", "description": "lunch", "date": "2024-01-15"}'
```

Response `201`:
```json
{
  "id": "...",
  "amount": 5000,
  "category": "food",
  "description": "lunch",
  "date": "2024-01-15T00:00:00Z",
  "created_at": "..."
}
```

Replay with same key → `200` (cached). Same key + different body → `409`.

### GET /expenses

```bash
# All expenses, newest first
curl http://localhost:8080/expenses

# Filtered by category
curl "http://localhost:8080/expenses?category=food"
```

Response `200`:
```json
{
  "expenses": [...]
}
```

## Project Structure

```
.
├── backend/
│   ├── cmd/server/main.go          # Entry point, DI wiring, graceful shutdown
│   └── internal/
│       ├── config/                 # Env-based config
│       ├── db/                     # pgx connection pool
│       ├── handler/                # HTTP handlers + response helpers
│       ├── middleware/             # RequestID, Logger, Timeout
│       ├── migrations/             # SQL migration files
│       ├── model/                  # Domain types + errors
│       ├── repository/             # PostgreSQL implementations
│       └── service/                # Business logic + idempotency
├── frontend/
│   ├── app/                        # Next.js App Router pages
│   ├── components/                 # ExpenseForm, ExpenseList, FilterBar
│   └── lib/                        # API client, types, money utils
├── docker-compose.yml
└── .env.example
```

## Tradeoffs

| Decision | Chosen | Alternative | Why |
|---|---|---|---|
| Money type | `int64` paise | `decimal` library | Zero deps, sufficient precision |
| ORM | Raw `pgx` | GORM / sqlc | Explicit SQL, no reflection overhead |
| Idempotency storage | DB table | Redis | No extra infra, ACID-safe |
| Frontend state | TanStack Query | Zustand + fetch | Server-state caching built-in |
