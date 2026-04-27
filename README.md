# Expense Tracker

A production-grade personal expense tracking application built as a full-stack system — not a throwaway prototype.

**Live:**
- Frontend: https://expense-tracker-eta-swart-72.vercel.app
- Backend API: https://expense-tracker-u1oc.onrender.com/health


## Architecture

```
Browser (Next.js / Vercel)
        │  HTTPS + CORS
        ▼
Go API (chi router / Render)
        │
        ├── handler      ← HTTP in/out, request parsing, response writing
        ├── middleware   ← Auth (JWT), RequestID, Logger, Timeout, Recoverer
        ├── service      ← Business logic, idempotency, validation
        └── repository   ← All SQL, pgx v5
                │
                ▼
        PostgreSQL (Render managed DB)
```

Clean layered architecture: each layer only imports the layer below it. Handlers know nothing about SQL; repositories know nothing about HTTP.

---

## Technology Choices

### Backend — Go + chi

**Why Go:** Compiled binary, low memory footprint (important for free-tier Render), strong standard library for HTTP servers, first-class concurrency primitives.

**Why chi:** Lightweight router that composes standard `net/http` middleware without wrapping it. No framework lock-in — every middleware is a plain `func(http.Handler) http.Handler`.

**Why pgx v5 (not GORM/sqlc):** Explicit SQL keeps queries readable and auditable. pgx is the fastest PostgreSQL driver for Go, with native support for pgx types (UUID, timestamptz). GORM adds reflection overhead and hides SQL; sqlc requires a code-gen step. For a small, stable schema, hand-written queries are better.

**Why golang-migrate:** Migration files are plain SQL — any DBA can read them. Version-controlled schema changes that run automatically on startup, no manual `psql` needed in CI or production.

**Why zap:** Structured JSON logging (key=value pairs) rather than printf-style. Every request gets a `request_id` that flows through all log lines, making production debugging possible.

### Frontend — Next.js 15 (App Router) + TanStack Query v5

**Why Next.js App Router:** Server Components reduce client-side bundle size. Built-in optimizations (image, font, prefetch). `'use client'` boundary is explicit — you know exactly what runs in the browser.

**Why TanStack Query:** Purpose-built for server-state: background refetch, stale-while-revalidate, automatic retries on network failure, and `queryKey`-based cache invalidation. This directly handles the "unreliable network / page refresh" requirement from the assignment without any manual state management.

### Database — PostgreSQL

**Why PostgreSQL over SQLite / in-memory:**
- ACID transactions for financial data — partial writes cannot corrupt the ledger
- `ON CONFLICT DO NOTHING` makes idempotency upserts safe under concurrent retries
- Native `gen_random_uuid()` (pgcrypto extension) for UUIDs
- Row-level isolation: each user's expenses are scoped by `user_id` FK
- Render provides a managed PostgreSQL instance with automated backups

---

## Key Design Decisions

### 1. Money as `int64` paise — never floats

```go
type Expense struct {
    Amount int64 `json:"amount"` // paise: 1 INR = 100 paise
}
```

IEEE 754 floating point cannot represent `0.1` exactly. `0.1 + 0.2 = 0.30000000000000004`. For money this means silent rounding errors that accumulate over time. Storing as integer paise eliminates this entirely. The frontend converts:
- **Input:** user types `100.50` → sent to API as `10050`
- **Display:** API returns `10050` → displayed as `₹100.50`

### 2. Idempotency for POST /expenses

The assignment explicitly requires correct behaviour when clients retry (network issues, double-clicks, page reloads). Implementation:

1. Client generates a UUID v4 `Idempotency-Key` per form session (stable across retries, rotated after successful submission)
2. Client sends `SHA-256(request body)` — wait, actually the **server** computes `SHA-256(request body)` on arrival
3. Backend checks the `idempotency_keys` table:
   - **Cache hit (same key + same body hash):** Return `200` with the cached response — no DB write
   - **Cache miss:** Create expense, persist `(key, body_hash, response_json)` atomically
   - **Conflict (same key, different body hash):** Return `409 IDEMPOTENCY_CONFLICT` — protects against accidental reuse

The `ON CONFLICT (key) DO NOTHING` write strategy makes concurrent retries safe without locks.

### 3. JWT Authentication with per-user data isolation

Every expense is owned by a user:

```sql
ALTER TABLE expenses ADD COLUMN user_id UUID NOT NULL REFERENCES users(id);
```

Every `GET /expenses` query includes `WHERE user_id = $1` using the JWT-extracted user ID. Users cannot see each other's data. The JWT is a 7-day `HS256` token signed with a `JWT_SECRET` environment variable (never hardcoded).

### 4. CORS configured via environment variable

```go
AllowedOrigins: []string{cfg.AllowedOrigins}, // set per-environment
AllowedHeaders: []string{"Accept", "Authorization", "Content-Type", "Idempotency-Key"},
```

`ALLOWED_ORIGINS` defaults to `*` locally, set to the exact Vercel URL in production. This prevents other origins from calling the API.

### 5. Graceful shutdown

```go
signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
<-quit
srv.Shutdown(ctx) // waits for in-flight requests (30s timeout)
```

Render sends `SIGTERM` before killing a container. The server drains existing connections before exiting, preventing dropped requests during deploys.

---

## Production Readiness Steps Taken

### Docker — multi-stage builds

```dockerfile
# Stage 1: compile
FROM golang:1.25-alpine AS builder
RUN go build -ldflags="-s -w" -o server ./cmd/server

# Stage 2: minimal runtime image
FROM alpine:3.19
COPY --from=builder /app/server .
COPY --from=builder /app/internal/migrations ./internal/migrations
```

Final image is ~20 MB (no Go toolchain, no build cache). `-ldflags="-s -w"` strips debug symbols and DWARF info. Frontend uses a 3-stage build (deps → build → standalone output) for the same reason.

### Health checks

```dockerfile
HEALTHCHECK CMD wget -qO- http://localhost:8080/health || exit 1
```

Docker Compose, Render, and container orchestrators use this to know when the backend is actually ready. The `frontend` and `backend` services won't start until their dependency passes the health check.

### Automated database backups

A dedicated `backup` sidecar service runs alongside the database:

```
scripts/backup.sh   → pg_dump → gzip → /backups/<db>_<timestamp>.sql.gz
scripts/restore.sh  → gunzip | psql (drop schema + reload)
backup/Dockerfile   → postgres:16-alpine + dcron
backup/entrypoint.sh → writes cron job, runs immediate backup on start, exec crond
```

- Schedule: configurable via `BACKUP_SCHEDULE` env var (default: `0 2 * * *` — 2 AM daily)
- Retention: configurable via `BACKUP_RETAIN_DAYS` (default: 7 days), older files pruned automatically
- Startup backup: runs immediately on container start to verify DB connectivity before the first scheduled job

### Backup restoration testing — GitHub Actions

`.github/workflows/backup-restore-test.yml` runs every Sunday:
1. Spins up a real PostgreSQL service container
2. Applies all migrations
3. Seeds sample data
4. Runs `backup.sh` to create a compressed dump
5. Runs `restore.sh` to reload from the dump
6. Verifies row count ≥ 1 — fails the workflow if data didn't survive the round-trip

This ensures backups are actually restorable, not just silently corrupt.

### CI — GitHub Actions

`.github/workflows/ci.yml` runs on every push to `main`:

**Backend job:**
- `go build ./...` — compilation check
- `go test ./internal/service/...` — unit tests (pure Go, no DB)
- `go test ./internal/repository/...` — integration tests (real PostgreSQL via Testcontainers)
- `golangci-lint run` — compiled from source with Go 1.25 to match `go.mod` (avoids pre-built binary version mismatch)

Linters enabled: `errcheck`, `gosimple`, `govet`, `ineffassign`, `staticcheck`, `unused`, `gofmt`, `misspell`

**Frontend job:**
- `tsc --noEmit` — TypeScript strict type-check
- `eslint` — Next.js lint rules
- `jest --ci` — component/unit tests
- `next build` — full production build

### Testing strategy

**Unit tests** (`internal/service/`): Test idempotency logic, hash computation, cache hit/miss/conflict behaviour. No database. Fast.

**Integration tests** (`internal/repository/`): Spin up a real `postgres:16-alpine` container via Testcontainers, create the full schema, run actual INSERT/SELECT queries. Tests verify:
- Expense creation and retrieval
- Category filtering
- Date sorting (ASC and DESC)
- User isolation — user A cannot see user B's expenses

### Deployed infrastructure

| Component | Platform | Notes |
|-----------|----------|-------|
| Frontend | Vercel | Auto-deploys on push to `main`; `NEXT_PUBLIC_API_URL` set to Render backend |
| Backend API | Render (Web Service) | Docker build from `backend/Dockerfile` |
| Database | Render (PostgreSQL) | Internal hostname for low-latency backend connection |

---

## How to Run Locally

**Prerequisites:** Docker + Docker Compose

```bash
cp .env.example .env
docker compose up --build
```

| Service | URL |
|---------|-----|
| Frontend | http://localhost:3000 |
| Backend API | http://localhost:8080 |
| Health | http://localhost:8080/health |

---

## API Reference

All authenticated endpoints require `Authorization: Bearer <token>`.

### POST /auth/register
```bash
curl -X POST http://localhost:8080/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email": "you@example.com", "password": "secret"}'
```

### POST /auth/login
```bash
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email": "you@example.com", "password": "secret"}'
# Returns: {"token": "<jwt>"}
```

### POST /expenses
Amount in **paise** (₹1 = 100 paise).

```bash
curl -X POST http://localhost:8080/expenses \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -H "Idempotency-Key: $(uuidgen)" \
  -d '{"amount": 5000, "category": "food", "description": "lunch", "date": "2024-01-15"}'
```

- Same key + same body → `200` (cached, no duplicate created)
- Same key + different body → `409 IDEMPOTENCY_CONFLICT`
- New key → `201 Created`

### GET /expenses

```bash
# All expenses, newest first (default)
curl http://localhost:8080/expenses -H "Authorization: Bearer <token>"

# Filter by category
curl "http://localhost:8080/expenses?category=food" -H "Authorization: Bearer <token>"

# Sort oldest first
curl "http://localhost:8080/expenses?sort=date_asc" -H "Authorization: Bearer <token>"
```

---

## Project Structure

```
.
├── backend/
│   ├── cmd/server/main.go           # Entry point: DI wiring, CORS, routes, graceful shutdown
│   └── internal/
│       ├── config/                  # Env-based config (DB, JWT, CORS, ports)
│       ├── db/                      # pgx connection pool
│       ├── handler/                 # HTTP handlers (thin — parse, call service, respond)
│       ├── middleware/              # JWT auth, RequestID, structured Logger, Timeout
│       ├── migrations/              # SQL migration files (golang-migrate format)
│       ├── model/                   # Domain types: Expense, User, errors
│       ├── repository/              # PostgreSQL: ExpenseRepo, UserRepo, IdempotencyRepo
│       └── service/                 # Business logic: ExpenseService, AuthService
├── frontend/
│   ├── app/                         # Next.js App Router pages (page.tsx, login, register)
│   ├── components/                  # ExpenseForm, ExpenseList, FilterBar, AuthGuard
│   └── lib/                         # API client, types, money utils, auth helpers
├── scripts/
│   ├── backup.sh                    # pg_dump → gzip with retention pruning
│   └── restore.sh                   # gunzip | psql restore
├── backup/
│   ├── Dockerfile                   # postgres:16-alpine + dcron sidecar
│   └── entrypoint.sh               # Writes cron tab, runs startup backup, starts crond
├── .github/workflows/
│   ├── ci.yml                       # Build + test + lint on every push
│   └── backup-restore-test.yml     # Weekly backup round-trip verification
├── docker-compose.yml               # postgres + backend + frontend + backup sidecar
└── .env.example                     # Template for local development
```

---

## Trade-offs

| Decision | Chosen | Alternative | Reason |
|----------|--------|-------------|--------|
| Money type | `int64` paise | `decimal` library | Zero deps, no rounding errors, sufficient for INR |
| ORM | Raw `pgx` queries | GORM / sqlc | Explicit SQL, no reflection, easy to audit |
| Idempotency store | PostgreSQL table | Redis | No extra infrastructure, ACID-safe, same RTO |
| Frontend state | TanStack Query | Zustand + fetch | Built-in retry, caching, background refetch for free |
| Auth | JWT (stateless) | Session cookies | No server-side session store needed; works cross-origin |
| Integration tests | Testcontainers | Mock repository | Tests real SQL against a real PostgreSQL — caught actual migration bugs |
| Migrations | golang-migrate | goose / atlas | Plain SQL files, minimal config, runs at startup |
| Backup | pg_dump sidecar | Managed backup only | Portable, testable, works with any PostgreSQL host |

### What I intentionally did not do

- **Pagination:** `GET /expenses` returns all rows. For a personal tracker this is fine; at scale, cursor-based pagination would be added.
- **Soft deletes / edit:** The assignment only specifies create and list. Delete/edit would add UI complexity without demonstrating additional design decisions.
- **Rate limiting:** Not required for a personal tool; trivial to add as chi middleware.
- **Full-text search on description:** Out of scope; a `tsvector` index + `to_tsquery` would be the PostgreSQL-native approach.
