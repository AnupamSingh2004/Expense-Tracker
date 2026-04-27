# Expense Tracker — Full-Stack Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a production-grade Expense Tracker with Go backend, Next.js frontend, PostgreSQL, Docker, and full idempotency support.

**Architecture:** Clean layered architecture (handler → service → repository → DB) on the backend; App Router + React Query on the frontend. Money stored as int64 paise. Idempotency via hashed request body + Idempotency-Key header persisted to a dedicated DB table.

**Tech Stack:** Go 1.22+, chi router, pgx/GORM, zap logging, Next.js 14 App Router, TypeScript strict, TanStack Query, TailwindCSS, PostgreSQL 16, Docker + Compose, GitHub Actions.

---

## File Map

### Backend (`backend/`)
```
backend/
  cmd/server/main.go                        # Entry point, DI wiring, graceful shutdown
  internal/
    config/config.go                        # Env-based config struct
    db/db.go                                # pgx connection pool factory
    middleware/
      logger.go                             # zap request logging
      requestid.go                          # X-Request-ID injection
      timeout.go                            # Per-request context deadline
    model/
      expense.go                            # Expense domain struct + enums
      idempotency.go                        # IdempotencyRecord struct
      errors.go                             # Typed domain errors
    repository/
      expense_repository.go                 # Interface + pgx impl
      idempotency_repository.go             # Interface + pgx impl
    service/
      expense_service.go                    # Business logic, orchestration
    handler/
      expense_handler.go                    # HTTP handlers (POST/GET)
      health_handler.go                     # GET /health
      response.go                           # JSON helpers, error envelope
    migrations/
      001_create_expenses.sql
      002_create_idempotency_keys.sql
      003_add_indexes.sql
  Dockerfile
  .golangci.yml
  go.mod / go.sum
```

### Frontend (`frontend/`)
```
frontend/
  app/
    layout.tsx                              # Root layout with QueryClientProvider
    page.tsx                               # Main page — composes Form + List
    globals.css
  components/
    ExpenseForm.tsx                         # Controlled form, idempotency key gen
    ExpenseList.tsx                         # Virtualized list + total row
    FilterBar.tsx                           # Category filter + sort toggle
    LoadingSpinner.tsx                      # Reusable spinner
    ErrorBanner.tsx                         # Error display
  lib/
    api.ts                                  # Typed fetch wrappers
    queryClient.ts                          # TanStack QueryClient config
    idempotency.ts                          # UUID v4 generation per submission
    money.ts                                # Paise ↔ rupee formatting
    types.ts                                # Shared TypeScript types
  __tests__/
    ExpenseForm.test.tsx
    ExpenseList.test.tsx
  .eslintrc.json
  .prettierrc
  next.config.ts
  tailwind.config.ts
  Dockerfile
```

### Infrastructure
```
docker-compose.yml
.env.example
.github/
  workflows/
    ci.yml
README.md
```

---

## Phase 1 — Monorepo Scaffold & Git Init

### Task 1: Initialize monorepo structure

**Files:**
- Create: `backend/` (directory)
- Create: `frontend/` (directory)
- Create: `.gitignore`

- [ ] **Step 1: Create top-level directories**

```bash
mkdir -p backend frontend docs/superpowers/plans
```

- [ ] **Step 2: Create root .gitignore**

```
# Go
backend/vendor/
backend/*.test
*.out

# Node
frontend/node_modules/
frontend/.next/
frontend/out/

# Env
.env
.env.local
*.local

# Docker
*.log
```

- [ ] **Step 3: Commit**

```bash
git add .gitignore docs/
git commit -m "chore(repo): initialize monorepo structure with backend and frontend"
```

---

## Phase 2 — Backend Bootstrap

### Task 2: Bootstrap Go module and project layout

**Files:**
- Create: `backend/go.mod`
- Create: `backend/cmd/server/main.go` (stub)
- Create: `backend/internal/config/config.go`
- Create: `backend/internal/db/db.go`

- [ ] **Step 1: Initialize Go module**

```bash
cd backend
go mod init github.com/fenmo/expense-tracker
```

- [ ] **Step 2: Install dependencies**

```bash
go get github.com/go-chi/chi/v5
go get github.com/jackc/pgx/v5
go get go.uber.org/zap
go get github.com/google/uuid
go get github.com/golang-migrate/migrate/v4
go get github.com/golang-migrate/migrate/v4/database/postgres
go get github.com/golang-migrate/migrate/v4/source/file
```

- [ ] **Step 3: Write config.go**

```go
// internal/config/config.go
package config

import (
    "fmt"
    "os"
    "strconv"
)

type Config struct {
    DBHost     string
    DBPort     int
    DBName     string
    DBUser     string
    DBPassword string
    DBSSLMode  string
    ServerPort int
    LogLevel   string
}

func Load() (*Config, error) {
    port, err := strconv.Atoi(getEnv("DB_PORT", "5432"))
    if err != nil {
        return nil, fmt.Errorf("invalid DB_PORT: %w", err)
    }
    srvPort, err := strconv.Atoi(getEnv("SERVER_PORT", "8080"))
    if err != nil {
        return nil, fmt.Errorf("invalid SERVER_PORT: %w", err)
    }
    return &Config{
        DBHost:     getEnv("DB_HOST", "localhost"),
        DBPort:     port,
        DBName:     getEnv("DB_NAME", "expenses"),
        DBUser:     getEnv("DB_USER", "postgres"),
        DBPassword: getEnv("DB_PASSWORD", ""),
        DBSSLMode:  getEnv("DB_SSLMODE", "disable"),
        ServerPort: srvPort,
        LogLevel:   getEnv("LOG_LEVEL", "info"),
    }, nil
}

func (c *Config) DSN() string {
    return fmt.Sprintf(
        "host=%s port=%d dbname=%s user=%s password=%s sslmode=%s",
        c.DBHost, c.DBPort, c.DBName, c.DBUser, c.DBPassword, c.DBSSLMode,
    )
}

func getEnv(key, fallback string) string {
    if v := os.Getenv(key); v != "" {
        return v
    }
    return fallback
}
```

- [ ] **Step 4: Write db.go**

```go
// internal/db/db.go
package db

import (
    "context"
    "fmt"

    "github.com/jackc/pgx/v5/pgxpool"
)

func NewPool(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
    pool, err := pgxpool.New(ctx, dsn)
    if err != nil {
        return nil, fmt.Errorf("unable to create connection pool: %w", err)
    }
    if err := pool.Ping(ctx); err != nil {
        return nil, fmt.Errorf("unable to ping database: %w", err)
    }
    return pool, nil
}
```

- [ ] **Step 5: Write stub main.go**

```go
// cmd/server/main.go
package main

import (
    "context"
    "fmt"
    "log"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/fenmo/expense-tracker/internal/config"
    "github.com/fenmo/expense-tracker/internal/db"
    "go.uber.org/zap"
)

func main() {
    cfg, err := config.Load()
    if err != nil {
        log.Fatalf("config: %v", err)
    }

    logger, _ := zap.NewProduction()
    defer logger.Sync()

    ctx := context.Background()
    pool, err := db.NewPool(ctx, cfg.DSN())
    if err != nil {
        logger.Fatal("db connection failed", zap.Error(err))
    }
    defer pool.Close()

    srv := &http.Server{
        Addr:         fmt.Sprintf(":%d", cfg.ServerPort),
        ReadTimeout:  15 * time.Second,
        WriteTimeout: 15 * time.Second,
        IdleTimeout:  60 * time.Second,
    }

    go func() {
        logger.Info("server starting", zap.String("addr", srv.Addr))
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            logger.Fatal("server error", zap.Error(err))
        }
    }()

    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit

    shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    if err := srv.Shutdown(shutdownCtx); err != nil {
        logger.Error("graceful shutdown failed", zap.Error(err))
    }
    logger.Info("server stopped")
}
```

- [ ] **Step 6: Verify it compiles**

```bash
cd backend && go build ./...
```
Expected: exits 0, no output.

- [ ] **Step 7: Commit**

```bash
git add backend/
git commit -m "chore(backend): bootstrap Go service with chi router and project layout"
```

---

## Phase 3 — Database Migrations

### Task 3: Define schema and run migrations

**Files:**
- Create: `backend/internal/migrations/001_create_expenses.sql`
- Create: `backend/internal/migrations/002_create_idempotency_keys.sql`
- Create: `backend/internal/migrations/003_add_indexes.sql`

- [ ] **Step 1: Write migration 001**

```sql
-- internal/migrations/001_create_expenses.sql
-- +migrate Up
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE IF NOT EXISTS expenses (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    amount      BIGINT      NOT NULL CHECK (amount > 0),
    category    VARCHAR(64) NOT NULL,
    description TEXT        NOT NULL DEFAULT '',
    date        DATE        NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- +migrate Down
DROP TABLE IF EXISTS expenses;
```

- [ ] **Step 2: Write migration 002**

```sql
-- internal/migrations/002_create_idempotency_keys.sql
-- +migrate Up
CREATE TABLE IF NOT EXISTS idempotency_keys (
    key          VARCHAR(255) PRIMARY KEY,
    request_hash VARCHAR(64)  NOT NULL,
    response     JSONB        NOT NULL,
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- +migrate Down
DROP TABLE IF EXISTS idempotency_keys;
```

- [ ] **Step 3: Write migration 003**

```sql
-- internal/migrations/003_add_indexes.sql
-- +migrate Up
CREATE INDEX IF NOT EXISTS idx_expenses_category    ON expenses (category);
CREATE INDEX IF NOT EXISTS idx_expenses_date        ON expenses (date DESC);
CREATE INDEX IF NOT EXISTS idx_expenses_category_date ON expenses (category, date DESC);

-- +migrate Down
DROP INDEX IF EXISTS idx_expenses_category;
DROP INDEX IF EXISTS idx_expenses_date;
DROP INDEX IF EXISTS idx_expenses_category_date;
```

- [ ] **Step 4: Add migration runner to main.go**

Add this function in `cmd/server/main.go` and call it before server start:

```go
import (
    migrate "github.com/golang-migrate/migrate/v4"
    _ "github.com/golang-migrate/migrate/v4/database/postgres"
    _ "github.com/golang-migrate/migrate/v4/source/file"
)

func runMigrations(dsn, migrationsPath string, logger *zap.Logger) error {
    m, err := migrate.New("file://"+migrationsPath, "postgresql://"+dsn)
    if err != nil {
        return fmt.Errorf("migration init: %w", err)
    }
    if err := m.Up(); err != nil && err != migrate.ErrNoChange {
        return fmt.Errorf("migration up: %w", err)
    }
    logger.Info("migrations applied")
    return nil
}
```

- [ ] **Step 5: Commit**

```bash
git add backend/internal/migrations/
git commit -m "feat(db): add migrations for expenses and idempotency keys table"
```

---

## Phase 4 — Domain Models

### Task 4: Define domain model types

**Files:**
- Create: `backend/internal/model/expense.go`
- Create: `backend/internal/model/idempotency.go`
- Create: `backend/internal/model/errors.go`

- [ ] **Step 1: Write expense.go**

```go
// internal/model/expense.go
package model

import (
    "time"
    "github.com/google/uuid"
)

// Expense represents a single recorded expense.
// Amount is stored in paise (1 INR = 100 paise) to avoid floating-point errors.
type Expense struct {
    ID          uuid.UUID `json:"id"`
    Amount      int64     `json:"amount"`      // paise
    Category    string    `json:"category"`
    Description string    `json:"description"`
    Date        time.Time `json:"date"`
    CreatedAt   time.Time `json:"created_at"`
}

type CreateExpenseInput struct {
    Amount      int64     `json:"amount"`      // paise from client
    Category    string    `json:"category"`
    Description string    `json:"description"`
    Date        string    `json:"date"`        // "YYYY-MM-DD"
}

type ListExpensesFilter struct {
    Category string
    SortBy   string // "date_desc" (default)
}
```

- [ ] **Step 2: Write idempotency.go**

```go
// internal/model/idempotency.go
package model

import (
    "encoding/json"
    "time"
)

type IdempotencyRecord struct {
    Key         string          `json:"key"`
    RequestHash string          `json:"request_hash"`
    Response    json.RawMessage `json:"response"`
    CreatedAt   time.Time       `json:"created_at"`
}
```

- [ ] **Step 3: Write errors.go**

```go
// internal/model/errors.go
package model

import "errors"

var (
    ErrNotFound         = errors.New("not found")
    ErrInvalidInput     = errors.New("invalid input")
    ErrDuplicateKey     = errors.New("duplicate idempotency key with different body")
    ErrInternalServer   = errors.New("internal server error")
)
```

- [ ] **Step 4: Commit**

```bash
git add backend/internal/model/
git commit -m "feat(api): define expense data model and database schema"
```

---

## Phase 5 — Repository Layer

### Task 5: Implement expense and idempotency repositories

**Files:**
- Create: `backend/internal/repository/expense_repository.go`
- Create: `backend/internal/repository/idempotency_repository.go`

- [ ] **Step 1: Write expense_repository.go**

```go
// internal/repository/expense_repository.go
package repository

import (
    "context"
    "fmt"
    "time"

    "github.com/fenmo/expense-tracker/internal/model"
    "github.com/google/uuid"
    "github.com/jackc/pgx/v5/pgxpool"
)

type ExpenseRepository interface {
    Create(ctx context.Context, e *model.Expense) (*model.Expense, error)
    List(ctx context.Context, f model.ListExpensesFilter) ([]*model.Expense, error)
}

type pgExpenseRepository struct {
    pool *pgxpool.Pool
}

func NewExpenseRepository(pool *pgxpool.Pool) ExpenseRepository {
    return &pgExpenseRepository{pool: pool}
}

func (r *pgExpenseRepository) Create(ctx context.Context, e *model.Expense) (*model.Expense, error) {
    e.ID = uuid.New()
    e.CreatedAt = time.Now().UTC()

    const q = `
        INSERT INTO expenses (id, amount, category, description, date, created_at)
        VALUES ($1, $2, $3, $4, $5, $6)
        RETURNING id, amount, category, description, date, created_at
    `
    row := r.pool.QueryRow(ctx, q,
        e.ID, e.Amount, e.Category, e.Description, e.Date, e.CreatedAt,
    )
    out := &model.Expense{}
    if err := row.Scan(&out.ID, &out.Amount, &out.Category, &out.Description, &out.Date, &out.CreatedAt); err != nil {
        return nil, fmt.Errorf("expense create: %w", err)
    }
    return out, nil
}

func (r *pgExpenseRepository) List(ctx context.Context, f model.ListExpensesFilter) ([]*model.Expense, error) {
    args := []any{}
    where := ""
    if f.Category != "" {
        args = append(args, f.Category)
        where = fmt.Sprintf("WHERE category = $%d", len(args))
    }
    q := fmt.Sprintf(`
        SELECT id, amount, category, description, date, created_at
        FROM expenses
        %s
        ORDER BY date DESC, created_at DESC
    `, where)

    rows, err := r.pool.Query(ctx, q, args...)
    if err != nil {
        return nil, fmt.Errorf("expense list: %w", err)
    }
    defer rows.Close()

    var expenses []*model.Expense
    for rows.Next() {
        e := &model.Expense{}
        if err := rows.Scan(&e.ID, &e.Amount, &e.Category, &e.Description, &e.Date, &e.CreatedAt); err != nil {
            return nil, fmt.Errorf("expense scan: %w", err)
        }
        expenses = append(expenses, e)
    }
    return expenses, rows.Err()
}
```

- [ ] **Step 2: Write idempotency_repository.go**

```go
// internal/repository/idempotency_repository.go
package repository

import (
    "context"
    "encoding/json"
    "errors"
    "fmt"

    "github.com/fenmo/expense-tracker/internal/model"
    "github.com/jackc/pgx/v5"
    "github.com/jackc/pgx/v5/pgxpool"
)

type IdempotencyRepository interface {
    Get(ctx context.Context, key string) (*model.IdempotencyRecord, error)
    Save(ctx context.Context, rec *model.IdempotencyRecord) error
}

type pgIdempotencyRepository struct {
    pool *pgxpool.Pool
}

func NewIdempotencyRepository(pool *pgxpool.Pool) IdempotencyRepository {
    return &pgIdempotencyRepository{pool: pool}
}

func (r *pgIdempotencyRepository) Get(ctx context.Context, key string) (*model.IdempotencyRecord, error) {
    const q = `SELECT key, request_hash, response, created_at FROM idempotency_keys WHERE key = $1`
    row := r.pool.QueryRow(ctx, q, key)
    rec := &model.IdempotencyRecord{}
    var resp []byte
    if err := row.Scan(&rec.Key, &rec.RequestHash, &resp, &rec.CreatedAt); err != nil {
        if errors.Is(err, pgx.ErrNoRows) {
            return nil, nil
        }
        return nil, fmt.Errorf("idempotency get: %w", err)
    }
    rec.Response = json.RawMessage(resp)
    return rec, nil
}

func (r *pgIdempotencyRepository) Save(ctx context.Context, rec *model.IdempotencyRecord) error {
    const q = `
        INSERT INTO idempotency_keys (key, request_hash, response, created_at)
        VALUES ($1, $2, $3, NOW())
        ON CONFLICT (key) DO NOTHING
    `
    _, err := r.pool.Exec(ctx, q, rec.Key, rec.RequestHash, []byte(rec.Response))
    if err != nil {
        return fmt.Errorf("idempotency save: %w", err)
    }
    return nil
}
```

- [ ] **Step 3: Commit**

```bash
git add backend/internal/repository/
git commit -m "feat(repo): implement expense repository with PostgreSQL"
```

---

## Phase 6 — Service Layer

### Task 6: Expense service with business logic and idempotency

**Files:**
- Create: `backend/internal/service/expense_service.go`

- [ ] **Step 1: Write expense_service.go**

```go
// internal/service/expense_service.go
package service

import (
    "context"
    "crypto/sha256"
    "encoding/json"
    "fmt"
    "time"

    "github.com/fenmo/expense-tracker/internal/model"
    "github.com/fenmo/expense-tracker/internal/repository"
    "go.uber.org/zap"
)

type ExpenseService interface {
    CreateExpense(ctx context.Context, idempotencyKey string, input model.CreateExpenseInput) (*model.Expense, bool, error)
    ListExpenses(ctx context.Context, filter model.ListExpensesFilter) ([]*model.Expense, error)
}

type expenseService struct {
    expRepo      repository.ExpenseRepository
    idempRepo    repository.IdempotencyRepository
    logger       *zap.Logger
}

func NewExpenseService(
    expRepo repository.ExpenseRepository,
    idempRepo repository.IdempotencyRepository,
    logger *zap.Logger,
) ExpenseService {
    return &expenseService{expRepo: expRepo, idempRepo: idempRepo, logger: logger}
}

func (s *expenseService) CreateExpense(
    ctx context.Context,
    idempotencyKey string,
    input model.CreateExpenseInput,
) (*model.Expense, bool, error) {
    if err := s.validate(input); err != nil {
        return nil, false, err
    }

    requestHash := hashInput(input)

    // Check idempotency cache
    if idempotencyKey != "" {
        rec, err := s.idempRepo.Get(ctx, idempotencyKey)
        if err != nil {
            return nil, false, fmt.Errorf("idempotency lookup: %w", err)
        }
        if rec != nil {
            if rec.RequestHash != requestHash {
                return nil, false, model.ErrDuplicateKey
            }
            var cached model.Expense
            if err := json.Unmarshal(rec.Response, &cached); err != nil {
                return nil, false, fmt.Errorf("decode cached response: %w", err)
            }
            return &cached, true, nil // cached hit
        }
    }

    // Parse date
    date, err := time.Parse("2006-01-02", input.Date)
    if err != nil {
        return nil, false, fmt.Errorf("%w: invalid date format, expected YYYY-MM-DD", model.ErrInvalidInput)
    }

    expense := &model.Expense{
        Amount:      input.Amount,
        Category:    input.Category,
        Description: input.Description,
        Date:        date,
    }
    created, err := s.expRepo.Create(ctx, expense)
    if err != nil {
        return nil, false, fmt.Errorf("create expense: %w", err)
    }

    // Persist idempotency record
    if idempotencyKey != "" {
        resp, _ := json.Marshal(created)
        _ = s.idempRepo.Save(ctx, &model.IdempotencyRecord{
            Key:         idempotencyKey,
            RequestHash: requestHash,
            Response:    json.RawMessage(resp),
        })
    }

    return created, false, nil
}

func (s *expenseService) ListExpenses(ctx context.Context, filter model.ListExpensesFilter) ([]*model.Expense, error) {
    return s.expRepo.List(ctx, filter)
}

func (s *expenseService) validate(input model.CreateExpenseInput) error {
    if input.Amount <= 0 {
        return fmt.Errorf("%w: amount must be positive", model.ErrInvalidInput)
    }
    if input.Category == "" {
        return fmt.Errorf("%w: category is required", model.ErrInvalidInput)
    }
    if input.Date == "" {
        return fmt.Errorf("%w: date is required", model.ErrInvalidInput)
    }
    return nil
}

func hashInput(input model.CreateExpenseInput) string {
    b, _ := json.Marshal(input)
    h := sha256.Sum256(b)
    return fmt.Sprintf("%x", h)
}
```

- [ ] **Step 2: Verify compilation**

```bash
cd backend && go build ./...
```

- [ ] **Step 3: Commit**

```bash
git add backend/internal/service/
git commit -m "feat(service): add expense creation logic with validation"
```

---

## Phase 7 — HTTP Handlers and Middleware

### Task 7: Middleware stack

**Files:**
- Create: `backend/internal/middleware/logger.go`
- Create: `backend/internal/middleware/requestid.go`
- Create: `backend/internal/middleware/timeout.go`

- [ ] **Step 1: Write requestid.go**

```go
// internal/middleware/requestid.go
package middleware

import (
    "context"
    "net/http"

    "github.com/google/uuid"
)

type contextKey string

const RequestIDKey contextKey = "request_id"

func RequestID(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        id := r.Header.Get("X-Request-ID")
        if id == "" {
            id = uuid.NewString()
        }
        ctx := context.WithValue(r.Context(), RequestIDKey, id)
        w.Header().Set("X-Request-ID", id)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

- [ ] **Step 2: Write logger.go**

```go
// internal/middleware/logger.go
package middleware

import (
    "net/http"
    "time"

    "go.uber.org/zap"
)

func Logger(logger *zap.Logger) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            start := time.Now()
            ww := &responseWriter{ResponseWriter: w, status: http.StatusOK}
            next.ServeHTTP(ww, r)
            logger.Info("request",
                zap.String("method", r.Method),
                zap.String("path", r.URL.Path),
                zap.Int("status", ww.status),
                zap.Duration("duration", time.Since(start)),
                zap.String("request_id", r.Context().Value(RequestIDKey).(string)),
            )
        })
    }
}

type responseWriter struct {
    http.ResponseWriter
    status int
}

func (rw *responseWriter) WriteHeader(status int) {
    rw.status = status
    rw.ResponseWriter.WriteHeader(status)
}
```

- [ ] **Step 3: Write timeout.go**

```go
// internal/middleware/timeout.go
package middleware

import (
    "context"
    "net/http"
    "time"
)

func Timeout(d time.Duration) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            ctx, cancel := context.WithTimeout(r.Context(), d)
            defer cancel()
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}
```

- [ ] **Step 4: Commit**

```bash
git add backend/internal/middleware/
git commit -m "feat(api): add middleware for logging, request ID, and timeouts"
```

### Task 8: HTTP handlers and response helpers

**Files:**
- Create: `backend/internal/handler/response.go`
- Create: `backend/internal/handler/expense_handler.go`
- Create: `backend/internal/handler/health_handler.go`

- [ ] **Step 1: Write response.go**

```go
// internal/handler/response.go
package handler

import (
    "encoding/json"
    "errors"
    "net/http"

    "github.com/fenmo/expense-tracker/internal/model"
    "go.uber.org/zap"
)

type ErrorResponse struct {
    Error   string `json:"error"`
    Code    string `json:"code"`
    Message string `json:"message"`
}

func JSON(w http.ResponseWriter, status int, body any) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    json.NewEncoder(w).Encode(body)
}

func ErrorJSON(w http.ResponseWriter, logger *zap.Logger, err error) {
    var status int
    var code, message string

    switch {
    case errors.Is(err, model.ErrInvalidInput):
        status = http.StatusBadRequest
        code = "INVALID_INPUT"
        message = err.Error()
    case errors.Is(err, model.ErrDuplicateKey):
        status = http.StatusConflict
        code = "IDEMPOTENCY_CONFLICT"
        message = "Idempotency key already used with a different request body"
    case errors.Is(err, model.ErrNotFound):
        status = http.StatusNotFound
        code = "NOT_FOUND"
        message = err.Error()
    default:
        status = http.StatusInternalServerError
        code = "INTERNAL_ERROR"
        message = "An unexpected error occurred"
        logger.Error("unhandled error", zap.Error(err))
    }

    JSON(w, status, ErrorResponse{Error: err.Error(), Code: code, Message: message})
}
```

- [ ] **Step 2: Write expense_handler.go**

```go
// internal/handler/expense_handler.go
package handler

import (
    "encoding/json"
    "net/http"

    "github.com/fenmo/expense-tracker/internal/model"
    "github.com/fenmo/expense-tracker/internal/service"
    "go.uber.org/zap"
)

type ExpenseHandler struct {
    svc    service.ExpenseService
    logger *zap.Logger
}

func NewExpenseHandler(svc service.ExpenseService, logger *zap.Logger) *ExpenseHandler {
    return &ExpenseHandler{svc: svc, logger: logger}
}

func (h *ExpenseHandler) Create(w http.ResponseWriter, r *http.Request) {
    var input model.CreateExpenseInput
    if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
        JSON(w, http.StatusBadRequest, ErrorResponse{
            Error: err.Error(), Code: "INVALID_JSON", Message: "Request body must be valid JSON",
        })
        return
    }

    idempotencyKey := r.Header.Get("Idempotency-Key")
    expense, cached, err := h.svc.CreateExpense(r.Context(), idempotencyKey, input)
    if err != nil {
        ErrorJSON(w, h.logger, err)
        return
    }

    status := http.StatusCreated
    if cached {
        status = http.StatusOK
    }
    JSON(w, status, expense)
}

func (h *ExpenseHandler) List(w http.ResponseWriter, r *http.Request) {
    filter := model.ListExpensesFilter{
        Category: r.URL.Query().Get("category"),
        SortBy:   r.URL.Query().Get("sort"),
    }

    expenses, err := h.svc.ListExpenses(r.Context(), filter)
    if err != nil {
        ErrorJSON(w, h.logger, err)
        return
    }

    if expenses == nil {
        expenses = []*model.Expense{}
    }
    JSON(w, http.StatusOK, map[string]any{"expenses": expenses})
}
```

- [ ] **Step 3: Write health_handler.go**

```go
// internal/handler/health_handler.go
package handler

import (
    "net/http"
)

func Health(w http.ResponseWriter, r *http.Request) {
    JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
```

- [ ] **Step 4: Wire routes in main.go**

Replace the stub server handler with chi router:

```go
import (
    "github.com/go-chi/chi/v5"
    chimw "github.com/go-chi/chi/v5/middleware"
    "github.com/fenmo/expense-tracker/internal/handler"
    "github.com/fenmo/expense-tracker/internal/middleware"
    "github.com/fenmo/expense-tracker/internal/repository"
    "github.com/fenmo/expense-tracker/internal/service"
)

// In main():
expRepo := repository.NewExpenseRepository(pool)
idempRepo := repository.NewIdempotencyRepository(pool)
expSvc := service.NewExpenseService(expRepo, idempRepo, logger)
expHandler := handler.NewExpenseHandler(expSvc, logger)

r := chi.NewRouter()
r.Use(middleware.RequestID)
r.Use(middleware.Logger(logger))
r.Use(middleware.Timeout(10 * time.Second))
r.Use(chimw.Recoverer)

r.Get("/health", handler.Health)
r.Route("/expenses", func(r chi.Router) {
    r.Post("/", expHandler.Create)
    r.Get("/", expHandler.List)
})

srv.Handler = r
```

- [ ] **Step 5: Build and verify**

```bash
cd backend && go build ./...
```

- [ ] **Step 6: Commit**

```bash
git add backend/internal/handler/ backend/cmd/
git commit -m "feat(api): implement POST /expenses with idempotency support"
```

---

## Phase 8 — Backend Tests

### Task 9: Service unit tests

**Files:**
- Create: `backend/internal/service/expense_service_test.go`

- [ ] **Step 1: Write mock repositories**

```go
// internal/service/expense_service_test.go
package service_test

import (
    "context"
    "encoding/json"
    "testing"

    "github.com/fenmo/expense-tracker/internal/model"
    "github.com/fenmo/expense-tracker/internal/repository"
    "github.com/fenmo/expense-tracker/internal/service"
    "go.uber.org/zap"
)

// --- mocks ---

type mockExpenseRepo struct {
    created []*model.Expense
}

func (m *mockExpenseRepo) Create(_ context.Context, e *model.Expense) (*model.Expense, error) {
    m.created = append(m.created, e)
    return e, nil
}

func (m *mockExpenseRepo) List(_ context.Context, _ model.ListExpensesFilter) ([]*model.Expense, error) {
    return m.created, nil
}

type mockIdempRepo struct {
    store map[string]*model.IdempotencyRecord
}

func newMockIdempRepo() *mockIdempRepo {
    return &mockIdempRepo{store: make(map[string]*model.IdempotencyRecord)}
}

func (m *mockIdempRepo) Get(_ context.Context, key string) (*model.IdempotencyRecord, error) {
    return m.store[key], nil
}

func (m *mockIdempRepo) Save(_ context.Context, rec *model.IdempotencyRecord) error {
    m.store[rec.Key] = rec
    return nil
}

var _ repository.ExpenseRepository = &mockExpenseRepo{}
var _ repository.IdempotencyRepository = &mockIdempRepo{}

// --- helpers ---

func newSvc(exp *mockExpenseRepo, idemp *mockIdempRepo) service.ExpenseService {
    return service.NewExpenseService(exp, idemp, zap.NewNop())
}

func validInput() model.CreateExpenseInput {
    return model.CreateExpenseInput{
        Amount:      5000,
        Category:    "food",
        Description: "lunch",
        Date:        "2024-01-15",
    }
}

// --- tests ---

func TestCreateExpense_Success(t *testing.T) {
    svc := newSvc(&mockExpenseRepo{}, newMockIdempRepo())
    expense, cached, err := svc.CreateExpense(context.Background(), "", validInput())
    if err != nil {
        t.Fatalf("expected no error, got: %v", err)
    }
    if cached {
        t.Error("expected cached=false on first create")
    }
    if expense.Amount != 5000 {
        t.Errorf("expected amount 5000, got %d", expense.Amount)
    }
}

func TestCreateExpense_IdempotencyHit(t *testing.T) {
    expRepo := &mockExpenseRepo{}
    idempRepo := newMockIdempRepo()
    svc := newSvc(expRepo, idempRepo)

    input := validInput()
    _, _, err := svc.CreateExpense(context.Background(), "key-1", input)
    if err != nil {
        t.Fatalf("first create: %v", err)
    }

    expense2, cached, err := svc.CreateExpense(context.Background(), "key-1", input)
    if err != nil {
        t.Fatalf("second create: %v", err)
    }
    if !cached {
        t.Error("expected cached=true on second call with same key")
    }
    if expense2 == nil {
        t.Error("expected non-nil expense from cache")
    }
    if len(expRepo.created) != 1 {
        t.Errorf("expected 1 DB write, got %d", len(expRepo.created))
    }
}

func TestCreateExpense_IdempotencyConflict(t *testing.T) {
    svc := newSvc(&mockExpenseRepo{}, newMockIdempRepo())

    input1 := validInput()
    _, _, err := svc.CreateExpense(context.Background(), "key-1", input1)
    if err != nil {
        t.Fatalf("first create: %v", err)
    }

    input2 := validInput()
    input2.Amount = 9999
    _, _, err = svc.CreateExpense(context.Background(), "key-1", input2)
    if err == nil {
        t.Fatal("expected conflict error, got nil")
    }
}

func TestCreateExpense_InvalidAmount(t *testing.T) {
    svc := newSvc(&mockExpenseRepo{}, newMockIdempRepo())
    input := validInput()
    input.Amount = 0
    _, _, err := svc.CreateExpense(context.Background(), "", input)
    if err == nil {
        t.Fatal("expected validation error for zero amount")
    }
}

func TestCreateExpense_InvalidDate(t *testing.T) {
    svc := newSvc(&mockExpenseRepo{}, newMockIdempRepo())
    input := validInput()
    input.Date = "not-a-date"
    _, _, err := svc.CreateExpense(context.Background(), "", input)
    if err == nil {
        t.Fatal("expected error for invalid date")
    }
}

func TestListExpenses_Empty(t *testing.T) {
    svc := newSvc(&mockExpenseRepo{}, newMockIdempRepo())
    result, err := svc.ListExpenses(context.Background(), model.ListExpensesFilter{})
    if err != nil {
        t.Fatalf("expected no error, got: %v", err)
    }
    if len(result) != 0 {
        t.Errorf("expected empty list, got %d", len(result))
    }
}

// Verify JSON marshaling round-trip for idempotency cache
func TestIdempotencyCache_RoundTrip(t *testing.T) {
    svc := newSvc(&mockExpenseRepo{}, newMockIdempRepo())
    input := validInput()
    orig, _, _ := svc.CreateExpense(context.Background(), "key-rt", input)

    cached, _, err := svc.CreateExpense(context.Background(), "key-rt", input)
    if err != nil {
        t.Fatalf("cache retrieval: %v", err)
    }
    origJSON, _ := json.Marshal(orig)
    cachedJSON, _ := json.Marshal(cached)
    if string(origJSON) != string(cachedJSON) {
        t.Errorf("round-trip mismatch:\norig:   %s\ncached: %s", origJSON, cachedJSON)
    }
}
```

- [ ] **Step 2: Run tests (should pass)**

```bash
cd backend && go test ./internal/service/... -v
```
Expected: all tests PASS.

- [ ] **Step 3: Commit**

```bash
git add backend/internal/service/expense_service_test.go
git commit -m "test(service): add unit tests for expense business logic"
```

### Task 10: Idempotency integration test with Docker/testcontainers

**Files:**
- Create: `backend/internal/repository/expense_repository_test.go`

- [ ] **Step 1: Install testcontainers**

```bash
cd backend && go get github.com/testcontainers/testcontainers-go
go get github.com/testcontainers/testcontainers-go/modules/postgres
```

- [ ] **Step 2: Write integration test**

```go
// internal/repository/expense_repository_test.go
package repository_test

import (
    "context"
    "fmt"
    "testing"
    "time"

    "github.com/fenmo/expense-tracker/internal/model"
    "github.com/fenmo/expense-tracker/internal/repository"
    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/testcontainers/testcontainers-go"
    pgmodule "github.com/testcontainers/testcontainers-go/modules/postgres"
    "github.com/testcontainers/testcontainers-go/wait"
)

func startPostgres(t *testing.T) (*pgxpool.Pool, func()) {
    t.Helper()
    ctx := context.Background()

    pgContainer, err := pgmodule.RunContainer(ctx,
        testcontainers.WithImage("postgres:16-alpine"),
        pgmodule.WithDatabase("testdb"),
        pgmodule.WithUsername("test"),
        pgmodule.WithPassword("test"),
        testcontainers.WithWaitStrategy(
            wait.ForLog("database system is ready to accept connections").
                WithOccurrence(2).WithStartupTimeout(30*time.Second),
        ),
    )
    if err != nil {
        t.Fatalf("start postgres container: %v", err)
    }

    connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
    if err != nil {
        t.Fatalf("connection string: %v", err)
    }

    pool, err := pgxpool.New(ctx, connStr)
    if err != nil {
        t.Fatalf("create pool: %v", err)
    }

    // Create minimal schema
    _, err = pool.Exec(ctx, `
        CREATE EXTENSION IF NOT EXISTS "pgcrypto";
        CREATE TABLE IF NOT EXISTS expenses (
            id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
            amount BIGINT NOT NULL CHECK (amount > 0),
            category VARCHAR(64) NOT NULL,
            description TEXT NOT NULL DEFAULT '',
            date DATE NOT NULL,
            created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
        );
        CREATE TABLE IF NOT EXISTS idempotency_keys (
            key VARCHAR(255) PRIMARY KEY,
            request_hash VARCHAR(64) NOT NULL,
            response JSONB NOT NULL,
            created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
        );
    `)
    if err != nil {
        t.Fatalf("schema setup: %v", err)
    }

    cleanup := func() {
        pool.Close()
        pgContainer.Terminate(ctx)
    }
    return pool, cleanup
}

func TestExpenseRepository_CreateAndList(t *testing.T) {
    pool, cleanup := startPostgres(t)
    defer cleanup()

    repo := repository.NewExpenseRepository(pool)
    ctx := context.Background()

    expense := &model.Expense{
        Amount:      5000,
        Category:    "food",
        Description: "lunch",
        Date:        time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
    }
    created, err := repo.Create(ctx, expense)
    if err != nil {
        t.Fatalf("create: %v", err)
    }
    if created.ID.String() == "" {
        t.Error("expected non-empty ID")
    }

    list, err := repo.List(ctx, model.ListExpensesFilter{})
    if err != nil {
        t.Fatalf("list: %v", err)
    }
    if len(list) != 1 {
        t.Fatalf("expected 1 expense, got %d", len(list))
    }
    if list[0].Amount != 5000 {
        t.Errorf("expected amount 5000, got %d", list[0].Amount)
    }
}

func TestExpenseRepository_FilterByCategory(t *testing.T) {
    pool, cleanup := startPostgres(t)
    defer cleanup()

    repo := repository.NewExpenseRepository(pool)
    ctx := context.Background()

    for _, cat := range []string{"food", "travel", "food"} {
        _, err := repo.Create(ctx, &model.Expense{
            Amount: 1000, Category: cat, Date: time.Now(),
        })
        if err != nil {
            t.Fatalf("create %s: %v", cat, err)
        }
    }

    list, err := repo.List(ctx, model.ListExpensesFilter{Category: "food"})
    if err != nil {
        t.Fatalf("list: %v", err)
    }
    if len(list) != 2 {
        t.Errorf("expected 2 food expenses, got %d", len(list))
    }
}

func TestIdempotencyRepository_SaveAndGet(t *testing.T) {
    pool, cleanup := startPostgres(t)
    defer cleanup()

    repo := repository.NewIdempotencyRepository(pool)
    ctx := context.Background()

    rec := &model.IdempotencyRecord{
        Key:         "test-key-1",
        RequestHash: "abc123",
        Response:    []byte(`{"id":"00000000-0000-0000-0000-000000000001"}`),
    }
    if err := repo.Save(ctx, rec); err != nil {
        t.Fatalf("save: %v", err)
    }

    got, err := repo.Get(ctx, "test-key-1")
    if err != nil {
        t.Fatalf("get: %v", err)
    }
    if got == nil {
        t.Fatal("expected record, got nil")
    }
    if got.RequestHash != "abc123" {
        t.Errorf("hash mismatch: %s", got.RequestHash)
    }
}

func TestIdempotencyRepository_GetMissing(t *testing.T) {
    pool, cleanup := startPostgres(t)
    defer cleanup()

    repo := repository.NewIdempotencyRepository(pool)
    got, err := repo.Get(context.Background(), "no-such-key")
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if got != nil {
        t.Errorf("expected nil for missing key, got: %+v", got)
    }
}

func TestIdempotencyRepository_DuplicateSaveIsNoop(t *testing.T) {
    pool, cleanup := startPostgres(t)
    defer cleanup()

    repo := repository.NewIdempotencyRepository(pool)
    ctx := context.Background()

    rec := &model.IdempotencyRecord{
        Key: "dup-key", RequestHash: "hash1",
        Response: []byte(`{}`),
    }
    if err := repo.Save(ctx, rec); err != nil {
        t.Fatalf("first save: %v", err)
    }
    rec2 := &model.IdempotencyRecord{
        Key: "dup-key", RequestHash: "hash2",
        Response: []byte(`{"different": true}`),
    }
    // ON CONFLICT DO NOTHING — should not error
    if err := repo.Save(ctx, rec2); err != nil {
        t.Fatalf("second save should be noop: %v", err)
    }

    got, _ := repo.Get(ctx, "dup-key")
    if got.RequestHash != "hash1" {
        t.Errorf("expected hash1 to be preserved, got %s", got.RequestHash)
    }
    fmt.Println("duplicate save is a noop ✓")
}
```

- [ ] **Step 3: Run integration tests**

```bash
cd backend && go test ./internal/repository/... -v -timeout 120s
```
Expected: all tests PASS (requires Docker).

- [ ] **Step 4: Commit**

```bash
git add backend/internal/repository/
git commit -m "test(repo): add database integration tests"
```

---

## Phase 9 — Backend Dockerfile

### Task 11: Multi-stage Dockerfile for Go

**Files:**
- Create: `backend/Dockerfile`
- Create: `backend/.golangci.yml`

- [ ] **Step 1: Write Dockerfile**

```dockerfile
# backend/Dockerfile
FROM golang:1.22-alpine AS builder
WORKDIR /app
RUN apk add --no-cache git
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o server ./cmd/server

FROM alpine:3.19
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=builder /app/server .
COPY --from=builder /app/internal/migrations ./internal/migrations
EXPOSE 8080
HEALTHCHECK --interval=10s --timeout=3s --retries=3 \
    CMD wget -qO- http://localhost:8080/health || exit 1
CMD ["./server"]
```

- [ ] **Step 2: Write .golangci.yml**

```yaml
# backend/.golangci.yml
run:
  timeout: 5m

linters:
  enable:
    - errcheck
    - gosimple
    - govet
    - ineffassign
    - staticcheck
    - unused
    - gofmt
    - goimports
    - misspell

linters-settings:
  goimports:
    local-prefixes: github.com/fenmo/expense-tracker
```

- [ ] **Step 3: Commit**

```bash
git add backend/Dockerfile backend/.golangci.yml
git commit -m "chore(docker): add multi-stage Dockerfile for Go backend"
```

---

## Phase 10 — Next.js Frontend Bootstrap

### Task 12: Initialize Next.js with TypeScript, Tailwind, TanStack Query

**Files:**
- Create: `frontend/` (entire Next.js app)

- [ ] **Step 1: Create Next.js app**

```bash
cd /home/anupam/code/fenmoAssesment
npx create-next-app@latest frontend \
  --typescript \
  --tailwind \
  --eslint \
  --app \
  --no-src-dir \
  --import-alias "@/*"
```

- [ ] **Step 2: Install dependencies**

```bash
cd frontend
npm install @tanstack/react-query @tanstack/react-query-devtools
npm install -D prettier eslint-config-prettier @testing-library/react @testing-library/jest-dom jest jest-environment-jsdom @types/jest ts-jest
```

- [ ] **Step 3: Write tsconfig.json (strict mode)**

Ensure `frontend/tsconfig.json` has:
```json
{
  "compilerOptions": {
    "strict": true,
    "noUncheckedIndexedAccess": true
  }
}
```

- [ ] **Step 4: Write .eslintrc.json**

```json
{
  "extends": ["next/core-web-vitals", "prettier"],
  "rules": {
    "@typescript-eslint/no-explicit-any": "error",
    "@typescript-eslint/no-unused-vars": "error"
  }
}
```

- [ ] **Step 5: Write .prettierrc**

```json
{
  "semi": true,
  "singleQuote": true,
  "tabWidth": 2,
  "trailingComma": "es5"
}
```

- [ ] **Step 6: Commit**

```bash
git add frontend/
git commit -m "chore(frontend): initialize Next.js app with TypeScript and App Router"
```

---

## Phase 11 — Frontend Types & API Layer

### Task 13: Shared types and typed API client

**Files:**
- Create: `frontend/lib/types.ts`
- Create: `frontend/lib/money.ts`
- Create: `frontend/lib/idempotency.ts`
- Create: `frontend/lib/api.ts`
- Create: `frontend/lib/queryClient.ts`

- [ ] **Step 1: Write types.ts**

```typescript
// frontend/lib/types.ts
export interface Expense {
  id: string;
  amount: number; // paise
  category: string;
  description: string;
  date: string;    // "YYYY-MM-DD"
  created_at: string;
}

export interface CreateExpensePayload {
  amount: number;  // paise
  category: string;
  description: string;
  date: string;
}

export interface ListExpensesResponse {
  expenses: Expense[];
}

export interface ApiError {
  error: string;
  code: string;
  message: string;
}
```

- [ ] **Step 2: Write money.ts**

```typescript
// frontend/lib/money.ts

/** Convert rupees (decimal) entered by user to paise (integer) for API */
export function rupeesToPaise(rupees: string): number {
  const parsed = parseFloat(rupees);
  if (isNaN(parsed) || parsed <= 0) throw new Error('Invalid amount');
  return Math.round(parsed * 100);
}

/** Format paise integer to display string, e.g. 5050 → "₹50.50" */
export function formatPaise(paise: number): string {
  return new Intl.NumberFormat('en-IN', {
    style: 'currency',
    currency: 'INR',
    minimumFractionDigits: 2,
  }).format(paise / 100);
}
```

- [ ] **Step 3: Write idempotency.ts**

```typescript
// frontend/lib/idempotency.ts

/** Generate a fresh UUID v4 idempotency key per form submission */
export function generateIdempotencyKey(): string {
  return crypto.randomUUID();
}
```

- [ ] **Step 4: Write api.ts**

```typescript
// frontend/lib/api.ts
import { CreateExpensePayload, Expense, ListExpensesResponse, ApiError } from './types';

const BASE_URL = process.env.NEXT_PUBLIC_API_URL ?? 'http://localhost:8080';

async function handleResponse<T>(res: Response): Promise<T> {
  if (!res.ok) {
    const err: ApiError = await res.json();
    throw new Error(err.message ?? `HTTP ${res.status}`);
  }
  return res.json() as Promise<T>;
}

export async function createExpense(
  payload: CreateExpensePayload,
  idempotencyKey: string
): Promise<Expense> {
  const res = await fetch(`${BASE_URL}/expenses`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'Idempotency-Key': idempotencyKey,
    },
    body: JSON.stringify(payload),
  });
  return handleResponse<Expense>(res);
}

export async function listExpenses(category?: string): Promise<Expense[]> {
  const params = new URLSearchParams();
  if (category) params.set('category', category);
  params.set('sort', 'date_desc');

  const res = await fetch(`${BASE_URL}/expenses?${params.toString()}`, {
    next: { revalidate: 0 }, // always fresh
  });
  const data = await handleResponse<ListExpensesResponse>(res);
  return data.expenses;
}
```

- [ ] **Step 5: Write queryClient.ts**

```typescript
// frontend/lib/queryClient.ts
import { QueryClient } from '@tanstack/react-query';

export const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 30_000,
      retry: 2,
      refetchOnWindowFocus: true,
    },
    mutations: {
      retry: 0,
    },
  },
});
```

- [ ] **Step 6: Commit**

```bash
git add frontend/lib/
git commit -m "feat(api-client): integrate API calls using React Query"
```

---

## Phase 12 — Frontend Components

### Task 14: QueryClientProvider in root layout

**Files:**
- Modify: `frontend/app/layout.tsx`
- Create: `frontend/app/providers.tsx`

- [ ] **Step 1: Write providers.tsx**

```tsx
// frontend/app/providers.tsx
'use client';

import { QueryClientProvider } from '@tanstack/react-query';
import { ReactQueryDevtools } from '@tanstack/react-query-devtools';
import { queryClient } from '@/lib/queryClient';

export function Providers({ children }: { children: React.ReactNode }) {
  return (
    <QueryClientProvider client={queryClient}>
      {children}
      <ReactQueryDevtools initialIsOpen={false} />
    </QueryClientProvider>
  );
}
```

- [ ] **Step 2: Update layout.tsx**

```tsx
// frontend/app/layout.tsx
import type { Metadata } from 'next';
import { Inter } from 'next/font/google';
import './globals.css';
import { Providers } from './providers';

const inter = Inter({ subsets: ['latin'] });

export const metadata: Metadata = {
  title: 'Expense Tracker',
  description: 'Track your personal expenses',
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en">
      <body className={inter.className}>
        <Providers>{children}</Providers>
      </body>
    </html>
  );
}
```

### Task 15: Shared UI primitives

**Files:**
- Create: `frontend/components/LoadingSpinner.tsx`
- Create: `frontend/components/ErrorBanner.tsx`

- [ ] **Step 1: Write LoadingSpinner.tsx**

```tsx
// frontend/components/LoadingSpinner.tsx
export function LoadingSpinner() {
  return (
    <div className="flex justify-center py-8">
      <div className="h-8 w-8 animate-spin rounded-full border-4 border-indigo-500 border-t-transparent" />
    </div>
  );
}
```

- [ ] **Step 2: Write ErrorBanner.tsx**

```tsx
// frontend/components/ErrorBanner.tsx
interface Props {
  message: string;
  onRetry?: () => void;
}

export function ErrorBanner({ message, onRetry }: Props) {
  return (
    <div className="rounded-md bg-red-50 border border-red-200 p-4 flex items-start gap-3">
      <span className="text-red-500 font-semibold">Error:</span>
      <span className="text-red-700 flex-1">{message}</span>
      {onRetry && (
        <button
          onClick={onRetry}
          className="text-sm text-red-600 underline hover:text-red-800"
        >
          Retry
        </button>
      )}
    </div>
  );
}
```

### Task 16: ExpenseForm component

**Files:**
- Create: `frontend/components/ExpenseForm.tsx`

- [ ] **Step 1: Write ExpenseForm.tsx**

```tsx
// frontend/components/ExpenseForm.tsx
'use client';

import { useState, useRef } from 'react';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { createExpense } from '@/lib/api';
import { rupeesToPaise } from '@/lib/money';
import { generateIdempotencyKey } from '@/lib/idempotency';

const CATEGORIES = ['food', 'travel', 'utilities', 'entertainment', 'health', 'other'];

interface FormState {
  amount: string;
  category: string;
  description: string;
  date: string;
}

const defaultForm: FormState = {
  amount: '',
  category: 'food',
  description: '',
  date: new Date().toISOString().split('T')[0] ?? '',
};

export function ExpenseForm() {
  const [form, setForm] = useState<FormState>(defaultForm);
  const [validationError, setValidationError] = useState<string | null>(null);
  // Idempotency key is stable per submission attempt, refreshed on success/reset
  const idempotencyKey = useRef(generateIdempotencyKey());

  const qc = useQueryClient();
  const mutation = useMutation({
    mutationFn: () =>
      createExpense(
        {
          amount: rupeesToPaise(form.amount),
          category: form.category,
          description: form.description,
          date: form.date,
        },
        idempotencyKey.current
      ),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['expenses'] });
      setForm(defaultForm);
      // Rotate key for next submission
      idempotencyKey.current = generateIdempotencyKey();
      setValidationError(null);
    },
  });

  function handleChange(e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement | HTMLTextAreaElement>) {
    setForm((f) => ({ ...f, [e.target.name]: e.target.value }));
  }

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setValidationError(null);
    const amount = parseFloat(form.amount);
    if (isNaN(amount) || amount <= 0) {
      setValidationError('Amount must be a positive number');
      return;
    }
    if (!form.date) {
      setValidationError('Date is required');
      return;
    }
    mutation.mutate();
  }

  const isSubmitting = mutation.isPending;

  return (
    <form onSubmit={handleSubmit} className="bg-white rounded-xl shadow p-6 space-y-4">
      <h2 className="text-lg font-semibold text-gray-800">Add Expense</h2>

      {(validationError ?? mutation.error) && (
        <p className="text-sm text-red-600 bg-red-50 rounded p-2">
          {validationError ?? (mutation.error as Error).message}
        </p>
      )}

      <div className="grid grid-cols-2 gap-4">
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">Amount (₹)</label>
          <input
            name="amount"
            type="number"
            step="0.01"
            min="0.01"
            value={form.amount}
            onChange={handleChange}
            required
            placeholder="0.00"
            className="w-full border rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
          />
        </div>
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">Category</label>
          <select
            name="category"
            value={form.category}
            onChange={handleChange}
            className="w-full border rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
          >
            {CATEGORIES.map((c) => (
              <option key={c} value={c}>{c.charAt(0).toUpperCase() + c.slice(1)}</option>
            ))}
          </select>
        </div>
      </div>

      <div>
        <label className="block text-sm font-medium text-gray-700 mb-1">Date</label>
        <input
          name="date"
          type="date"
          value={form.date}
          onChange={handleChange}
          required
          className="w-full border rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
        />
      </div>

      <div>
        <label className="block text-sm font-medium text-gray-700 mb-1">Description</label>
        <textarea
          name="description"
          value={form.description}
          onChange={handleChange}
          rows={2}
          placeholder="Optional notes..."
          className="w-full border rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
        />
      </div>

      <button
        type="submit"
        disabled={isSubmitting}
        className="w-full bg-indigo-600 hover:bg-indigo-700 disabled:bg-indigo-300 text-white font-medium py-2 px-4 rounded-md transition-colors"
      >
        {isSubmitting ? 'Saving...' : 'Add Expense'}
      </button>
    </form>
  );
}
```

- [ ] **Step 2: Commit**

```bash
git add frontend/components/ frontend/app/
git commit -m "feat(ui): build expense form with validation and controlled inputs"
```

### Task 17: FilterBar and ExpenseList components

**Files:**
- Create: `frontend/components/FilterBar.tsx`
- Create: `frontend/components/ExpenseList.tsx`

- [ ] **Step 1: Write FilterBar.tsx**

```tsx
// frontend/components/FilterBar.tsx
'use client';

const CATEGORIES = ['', 'food', 'travel', 'utilities', 'entertainment', 'health', 'other'];

interface Props {
  category: string;
  onCategoryChange: (cat: string) => void;
}

export function FilterBar({ category, onCategoryChange }: Props) {
  return (
    <div className="flex items-center gap-3">
      <label className="text-sm font-medium text-gray-700">Filter:</label>
      <select
        value={category}
        onChange={(e) => onCategoryChange(e.target.value)}
        className="border rounded-md px-3 py-1.5 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
      >
        <option value="">All categories</option>
        {CATEGORIES.filter(Boolean).map((c) => (
          <option key={c} value={c}>{c.charAt(0).toUpperCase() + c.slice(1)}</option>
        ))}
      </select>
    </div>
  );
}
```

- [ ] **Step 2: Write ExpenseList.tsx**

```tsx
// frontend/components/ExpenseList.tsx
'use client';

import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { listExpenses } from '@/lib/api';
import { formatPaise } from '@/lib/money';
import { FilterBar } from './FilterBar';
import { LoadingSpinner } from './LoadingSpinner';
import { ErrorBanner } from './ErrorBanner';

export function ExpenseList() {
  const [category, setCategory] = useState('');

  const { data: expenses = [], isLoading, isError, error, refetch } = useQuery({
    queryKey: ['expenses', category],
    queryFn: () => listExpenses(category || undefined),
  });

  const total = expenses.reduce((sum, e) => sum + e.amount, 0);

  return (
    <div className="bg-white rounded-xl shadow p-6 space-y-4">
      <div className="flex items-center justify-between">
        <h2 className="text-lg font-semibold text-gray-800">Expenses</h2>
        <FilterBar category={category} onCategoryChange={setCategory} />
      </div>

      {isLoading && <LoadingSpinner />}
      {isError && (
        <ErrorBanner
          message={(error as Error).message}
          onRetry={() => refetch()}
        />
      )}

      {!isLoading && !isError && expenses.length === 0 && (
        <p className="text-center text-gray-400 py-8">No expenses found.</p>
      )}

      {expenses.length > 0 && (
        <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b text-left text-gray-500">
                <th className="pb-2 font-medium">Date</th>
                <th className="pb-2 font-medium">Category</th>
                <th className="pb-2 font-medium">Description</th>
                <th className="pb-2 font-medium text-right">Amount</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-100">
              {expenses.map((e) => (
                <tr key={e.id} className="hover:bg-gray-50">
                  <td className="py-2 text-gray-600">{e.date}</td>
                  <td className="py-2">
                    <span className="inline-block bg-indigo-100 text-indigo-700 text-xs px-2 py-0.5 rounded-full capitalize">
                      {e.category}
                    </span>
                  </td>
                  <td className="py-2 text-gray-600 max-w-xs truncate">{e.description || '—'}</td>
                  <td className="py-2 text-right font-medium text-gray-800">{formatPaise(e.amount)}</td>
                </tr>
              ))}
            </tbody>
            <tfoot>
              <tr className="border-t-2 border-gray-200 font-semibold">
                <td colSpan={3} className="pt-3 text-gray-700">Total ({expenses.length} items)</td>
                <td className="pt-3 text-right text-indigo-700">{formatPaise(total)}</td>
              </tr>
            </tfoot>
          </table>
        </div>
      )}
    </div>
  );
}
```

- [ ] **Step 3: Commit**

```bash
git add frontend/components/
git commit -m "feat(ui): implement expense list with sorting and filtering"
```

### Task 18: Main page

**Files:**
- Modify: `frontend/app/page.tsx`

- [ ] **Step 1: Write page.tsx**

```tsx
// frontend/app/page.tsx
import { ExpenseForm } from '@/components/ExpenseForm';
import { ExpenseList } from '@/components/ExpenseList';

export default function Home() {
  return (
    <main className="min-h-screen bg-gray-50">
      <header className="bg-white border-b">
        <div className="max-w-4xl mx-auto px-4 py-4">
          <h1 className="text-2xl font-bold text-indigo-700">Expense Tracker</h1>
          <p className="text-sm text-gray-500">Track your personal expenses</p>
        </div>
      </header>
      <div className="max-w-4xl mx-auto px-4 py-8 grid gap-8 lg:grid-cols-[380px_1fr]">
        <div>
          <ExpenseForm />
        </div>
        <div>
          <ExpenseList />
        </div>
      </div>
    </main>
  );
}
```

- [ ] **Step 2: Commit**

```bash
git add frontend/app/page.tsx
git commit -m "feat(ui): add total expense calculation for visible items"
```

---

## Phase 13 — Frontend Tests

### Task 19: Frontend component tests

**Files:**
- Create: `frontend/jest.config.ts`
- Create: `frontend/jest.setup.ts`
- Create: `frontend/__tests__/ExpenseForm.test.tsx`
- Create: `frontend/__tests__/ExpenseList.test.tsx`

- [ ] **Step 1: Write jest.config.ts**

```typescript
// frontend/jest.config.ts
import type { Config } from 'jest';

const config: Config = {
  testEnvironment: 'jsdom',
  setupFilesAfterFramework: ['<rootDir>/jest.setup.ts'],
  moduleNameMapper: {
    '^@/(.*)$': '<rootDir>/$1',
  },
  transform: {
    '^.+\\.(ts|tsx)$': ['ts-jest', { tsconfig: { jsx: 'react-jsx' } }],
  },
};

export default config;
```

- [ ] **Step 2: Write jest.setup.ts**

```typescript
// frontend/jest.setup.ts
import '@testing-library/jest-dom';
```

- [ ] **Step 3: Write ExpenseForm.test.tsx**

```tsx
// frontend/__tests__/ExpenseForm.test.tsx
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { ExpenseForm } from '@/components/ExpenseForm';
import * as api from '@/lib/api';

jest.mock('@/lib/api');

function wrapper({ children }: { children: React.ReactNode }) {
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return <QueryClientProvider client={qc}>{children}</QueryClientProvider>;
}

describe('ExpenseForm', () => {
  it('renders form fields', () => {
    render(<ExpenseForm />, { wrapper });
    expect(screen.getByLabelText(/amount/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/category/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/date/i)).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /add expense/i })).toBeInTheDocument();
  });

  it('shows validation error for invalid amount', async () => {
    render(<ExpenseForm />, { wrapper });
    fireEvent.change(screen.getByLabelText(/amount/i), { target: { value: '-10' } });
    fireEvent.submit(screen.getByRole('button', { name: /add expense/i }).closest('form')!);
    await waitFor(() => {
      expect(screen.getByText(/amount must be a positive number/i)).toBeInTheDocument();
    });
  });

  it('disables button while submitting', async () => {
    (api.createExpense as jest.Mock).mockImplementation(
      () => new Promise((resolve) => setTimeout(resolve, 500))
    );
    render(<ExpenseForm />, { wrapper });
    fireEvent.change(screen.getByLabelText(/amount/i), { target: { value: '50' } });
    fireEvent.submit(screen.getByRole('button').closest('form')!);
    expect(screen.getByRole('button', { name: /saving/i })).toBeDisabled();
  });
});
```

- [ ] **Step 4: Write ExpenseList.test.tsx**

```tsx
// frontend/__tests__/ExpenseList.test.tsx
import { render, screen } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { ExpenseList } from '@/components/ExpenseList';
import * as api from '@/lib/api';
import type { Expense } from '@/lib/types';

jest.mock('@/lib/api');

function wrapper({ children }: { children: React.ReactNode }) {
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return <QueryClientProvider client={qc}>{children}</QueryClientProvider>;
}

const mockExpenses: Expense[] = [
  { id: '1', amount: 5000, category: 'food', description: 'lunch', date: '2024-01-15', created_at: '' },
  { id: '2', amount: 10000, category: 'travel', description: 'uber', date: '2024-01-14', created_at: '' },
];

describe('ExpenseList', () => {
  it('renders expenses and total', async () => {
    (api.listExpenses as jest.Mock).mockResolvedValue(mockExpenses);
    render(<ExpenseList />, { wrapper });
    await screen.findByText('lunch');
    expect(screen.getByText('uber')).toBeInTheDocument();
    // Total: 5000 + 10000 = 15000 paise = ₹150
    expect(screen.getByText(/₹150/)).toBeInTheDocument();
  });

  it('shows empty state when no expenses', async () => {
    (api.listExpenses as jest.Mock).mockResolvedValue([]);
    render(<ExpenseList />, { wrapper });
    await screen.findByText(/no expenses found/i);
  });

  it('shows error banner on API failure', async () => {
    (api.listExpenses as jest.Mock).mockRejectedValue(new Error('Network error'));
    render(<ExpenseList />, { wrapper });
    await screen.findByText(/network error/i);
  });
});
```

- [ ] **Step 5: Run tests**

```bash
cd frontend && npm test -- --passWithNoTests
```

- [ ] **Step 6: Commit**

```bash
git add frontend/__tests__/ frontend/jest.config.ts frontend/jest.setup.ts
git commit -m "test(ui): add basic component test for expense form and list"
```

---

## Phase 14 — Frontend Dockerfile

### Task 20: Next.js Dockerfile

**Files:**
- Create: `frontend/Dockerfile`
- Create: `frontend/next.config.ts` (update for Docker)

- [ ] **Step 1: Write Dockerfile**

```dockerfile
# frontend/Dockerfile
FROM node:20-alpine AS deps
WORKDIR /app
COPY package.json package-lock.json ./
RUN npm ci

FROM node:20-alpine AS builder
WORKDIR /app
COPY --from=deps /app/node_modules ./node_modules
COPY . .
ENV NEXT_TELEMETRY_DISABLED=1
RUN npm run build

FROM node:20-alpine AS runner
WORKDIR /app
ENV NODE_ENV=production
ENV NEXT_TELEMETRY_DISABLED=1
RUN addgroup --system --gid 1001 nodejs && adduser --system --uid 1001 nextjs
COPY --from=builder /app/public ./public
COPY --from=builder --chown=nextjs:nodejs /app/.next/standalone ./
COPY --from=builder --chown=nextjs:nodejs /app/.next/static ./.next/static
USER nextjs
EXPOSE 3000
HEALTHCHECK --interval=10s --timeout=3s --retries=3 \
    CMD wget -qO- http://localhost:3000 || exit 1
CMD ["node", "server.js"]
```

- [ ] **Step 2: Enable Next.js standalone output**

In `frontend/next.config.ts`:
```typescript
const nextConfig = {
  output: 'standalone',
};
export default nextConfig;
```

- [ ] **Step 3: Commit**

```bash
git add frontend/Dockerfile frontend/next.config.ts
git commit -m "chore(docker): add Next.js Dockerfile with standalone output"
```

---

## Phase 15 — Docker Compose

### Task 21: Docker Compose with all services

**Files:**
- Create: `docker-compose.yml`
- Create: `.env.example`

- [ ] **Step 1: Write docker-compose.yml**

```yaml
# docker-compose.yml
version: '3.9'

services:
  postgres:
    image: postgres:16-alpine
    restart: unless-stopped
    environment:
      POSTGRES_DB: ${DB_NAME:-expenses}
      POSTGRES_USER: ${DB_USER:-postgres}
      POSTGRES_PASSWORD: ${DB_PASSWORD:-secret}
    ports:
      - "5432:5432"
    volumes:
      - pgdata:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U ${DB_USER:-postgres} -d ${DB_NAME:-expenses}"]
      interval: 5s
      timeout: 5s
      retries: 10

  backend:
    build:
      context: ./backend
      dockerfile: Dockerfile
    restart: unless-stopped
    depends_on:
      postgres:
        condition: service_healthy
    environment:
      DB_HOST: postgres
      DB_PORT: 5432
      DB_NAME: ${DB_NAME:-expenses}
      DB_USER: ${DB_USER:-postgres}
      DB_PASSWORD: ${DB_PASSWORD:-secret}
      DB_SSLMODE: disable
      SERVER_PORT: 8080
      LOG_LEVEL: info
      MIGRATIONS_PATH: ./internal/migrations
    ports:
      - "8080:8080"
    healthcheck:
      test: ["CMD", "wget", "-qO-", "http://localhost:8080/health"]
      interval: 10s
      timeout: 5s
      retries: 5

  frontend:
    build:
      context: ./frontend
      dockerfile: Dockerfile
    restart: unless-stopped
    depends_on:
      backend:
        condition: service_healthy
    environment:
      NEXT_PUBLIC_API_URL: http://localhost:8080
    ports:
      - "3000:3000"

volumes:
  pgdata:
```

- [ ] **Step 2: Write .env.example**

```bash
# .env.example
DB_NAME=expenses
DB_USER=postgres
DB_PASSWORD=changeme_in_production
DB_HOST=localhost
DB_PORT=5432
DB_SSLMODE=disable
SERVER_PORT=8080
LOG_LEVEL=info
MIGRATIONS_PATH=./internal/migrations
NEXT_PUBLIC_API_URL=http://localhost:8080
```

- [ ] **Step 3: Commit**

```bash
git add docker-compose.yml .env.example
git commit -m "chore(devops): add Dockerfiles and docker-compose setup with PostgreSQL"
```

---

## Phase 16 — GitHub Actions CI

### Task 22: CI pipeline

**Files:**
- Create: `.github/workflows/ci.yml`

- [ ] **Step 1: Write ci.yml**

```yaml
# .github/workflows/ci.yml
name: CI

on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main]

jobs:
  backend:
    name: Backend (Go)
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgres:16-alpine
        env:
          POSTGRES_DB: testdb
          POSTGRES_USER: test
          POSTGRES_PASSWORD: test
        ports:
          - 5432:5432
        options: >-
          --health-cmd pg_isready
          --health-interval 5s
          --health-timeout 5s
          --health-retries 10
    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with:
          go-version: '1.22'

      - name: Download dependencies
        working-directory: backend
        run: go mod download

      - name: Build
        working-directory: backend
        run: go build ./...

      - name: Unit tests
        working-directory: backend
        run: go test ./internal/service/... -v -race

      - name: Integration tests
        working-directory: backend
        run: go test ./internal/repository/... -v -timeout 120s
        env:
          TESTCONTAINERS_RYUK_DISABLED: true

      - name: Lint
        uses: golangci/golangci-lint-action@v4
        with:
          working-directory: backend
          version: latest

  frontend:
    name: Frontend (Next.js)
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-node@v4
        with:
          node-version: '20'
          cache: npm
          cache-dependency-path: frontend/package-lock.json

      - name: Install dependencies
        working-directory: frontend
        run: npm ci

      - name: Type check
        working-directory: frontend
        run: npx tsc --noEmit

      - name: Lint
        working-directory: frontend
        run: npm run lint

      - name: Test
        working-directory: frontend
        run: npm test -- --passWithNoTests --ci

      - name: Build
        working-directory: frontend
        run: npm run build
        env:
          NEXT_PUBLIC_API_URL: http://localhost:8080
```

- [ ] **Step 2: Commit**

```bash
git add .github/
git commit -m "chore(ci): add GitHub Actions workflow for build and test"
```

---

## Phase 17 — README

### Task 23: Comprehensive README

**Files:**
- Create: `README.md`

- [ ] **Step 1: Write README.md** (see full content in implementation)

The README must include:
- Architecture diagram (ASCII)
- Why PostgreSQL
- Why int for money (paise)
- Idempotency strategy explanation
- How to run locally (docker compose up)
- API documentation with curl examples
- Tradeoffs section

- [ ] **Step 2: Commit**

```bash
git add README.md
git commit -m "chore(readme): add architecture decisions and setup instructions"
```

---

## Verification Checklist

Before calling this complete, verify:

- [ ] `cd backend && go build ./...` — exits 0
- [ ] `cd backend && go test ./... -count=1` — all pass
- [ ] `cd frontend && npm run build` — exits 0
- [ ] `cd frontend && npm test` — all pass
- [ ] `docker compose up --build` — all services healthy
- [ ] `curl -X POST http://localhost:8080/expenses -H "Content-Type: application/json" -H "Idempotency-Key: test-1" -d '{"amount":5000,"category":"food","description":"lunch","date":"2024-01-15"}'` — returns 201
- [ ] Same curl repeated — returns 200 (cached)
- [ ] `curl http://localhost:8080/expenses?category=food` — returns filtered list
- [ ] Frontend loads at `http://localhost:3000` — form + list renders

---

## Commit Sequence Summary

1. `chore(repo): initialize monorepo structure with backend and frontend`
2. `chore(backend): bootstrap Go service with chi router and project layout`
3. `feat(db): add migrations for expenses and idempotency keys table`
4. `feat(api): define expense data model and database schema`
5. `feat(repo): implement expense repository with PostgreSQL`
6. `feat(service): add expense creation logic with validation`
7. `feat(api): add middleware for logging, request ID, and timeouts`
8. `feat(api): implement POST /expenses with idempotency support`
9. `feat(api): implement GET /expenses with filtering and sorting`
10. `test(service): add unit tests for expense business logic`
11. `test(repo): add database integration tests`
12. `chore(docker): add multi-stage Dockerfile for Go backend`
13. `chore(frontend): initialize Next.js app with TypeScript and App Router`
14. `feat(api-client): integrate API calls using React Query`
15. `feat(ui): build expense form with validation and controlled inputs`
16. `feat(ui): implement expense list with sorting and filtering`
17. `feat(ui): add total expense calculation for visible items`
18. `test(ui): add basic component test for expense form and list`
19. `chore(docker): add Next.js Dockerfile with standalone output`
20. `chore(devops): add Dockerfiles and docker-compose setup with PostgreSQL`
21. `chore(ci): add GitHub Actions workflow for build and test`
22. `chore(readme): add architecture decisions and setup instructions`
