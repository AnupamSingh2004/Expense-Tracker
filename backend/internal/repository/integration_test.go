package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/fenmo/expense-tracker/internal/model"
	"github.com/fenmo/expense-tracker/internal/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	pgmodule "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func startDB(t *testing.T) (*pgxpool.Pool, func()) {
	t.Helper()
	ctx := context.Background()

	ctr, err := pgmodule.Run(ctx,
		"postgres:16-alpine",
		pgmodule.WithDatabase("testdb"),
		pgmodule.WithUsername("test"),
		pgmodule.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).WithStartupTimeout(60*time.Second),
		),
	)
	if err != nil {
		t.Fatalf("start postgres: %v", err)
	}

	connStr, err := ctr.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("connection string: %v", err)
	}

	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		t.Fatalf("create pool: %v", err)
	}

	_, err = pool.Exec(ctx, `
		CREATE EXTENSION IF NOT EXISTS "pgcrypto";

		CREATE TABLE IF NOT EXISTS users (
			id           UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
			email        VARCHAR(255) NOT NULL UNIQUE,
			password_hash TEXT        NOT NULL,
			created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);

		CREATE TABLE IF NOT EXISTS expenses (
			id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id     UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			amount      BIGINT      NOT NULL CHECK (amount > 0),
			category    VARCHAR(64) NOT NULL,
			description TEXT        NOT NULL DEFAULT '',
			date        DATE        NOT NULL,
			created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);

		CREATE TABLE IF NOT EXISTS idempotency_keys (
			key          VARCHAR(255) PRIMARY KEY,
			request_hash VARCHAR(64)  NOT NULL,
			response     JSONB        NOT NULL,
			created_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW()
		);
	`)
	if err != nil {
		t.Fatalf("schema setup: %v", err)
	}

	return pool, func() {
		pool.Close()
		ctr.Terminate(ctx) //nolint:errcheck
	}
}

// seedUser inserts a test user and returns their ID.
func seedUser(t *testing.T, pool *pgxpool.Pool) uuid.UUID {
	t.Helper()
	var id uuid.UUID
	err := pool.QueryRow(context.Background(),
		`INSERT INTO users (email, password_hash) VALUES ($1, $2) RETURNING id`,
		"test@example.com", "hash",
	).Scan(&id)
	if err != nil {
		t.Fatalf("seed user: %v", err)
	}
	return id
}

func TestExpenseRepo_CreateAndList(t *testing.T) {
	pool, cleanup := startDB(t)
	defer cleanup()

	userID := seedUser(t, pool)
	repo := repository.NewExpenseRepository(pool)
	ctx := context.Background()

	created, err := repo.Create(ctx, &model.Expense{
		UserID:      userID,
		Amount:      5000,
		Category:    "food",
		Description: "lunch",
		Date:        time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if created.ID.String() == "" {
		t.Error("expected non-empty ID")
	}

	list, err := repo.List(ctx, model.ListExpensesFilter{UserID: userID})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(list) != 1 || list[0].Amount != 5000 {
		t.Errorf("unexpected list: %+v", list)
	}
}

func TestExpenseRepo_FilterByCategory(t *testing.T) {
	pool, cleanup := startDB(t)
	defer cleanup()

	userID := seedUser(t, pool)
	repo := repository.NewExpenseRepository(pool)
	ctx := context.Background()

	for _, cat := range []string{"food", "travel", "food"} {
		_, err := repo.Create(ctx, &model.Expense{
			UserID: userID, Amount: 1000, Category: cat, Date: time.Now(),
		})
		if err != nil {
			t.Fatalf("create %s: %v", cat, err)
		}
	}

	list, err := repo.List(ctx, model.ListExpensesFilter{UserID: userID, Category: "food"})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(list) != 2 {
		t.Errorf("expected 2 food items, got %d", len(list))
	}
}

func TestExpenseRepo_IsolatedByUser(t *testing.T) {
	pool, cleanup := startDB(t)
	defer cleanup()

	repo := repository.NewExpenseRepository(pool)
	ctx := context.Background()

	// Create two different users
	var userA, userB uuid.UUID
	pool.QueryRow(ctx, `INSERT INTO users (email, password_hash) VALUES ('a@test.com','h') RETURNING id`).Scan(&userA) //nolint:errcheck
	pool.QueryRow(ctx, `INSERT INTO users (email, password_hash) VALUES ('b@test.com','h') RETURNING id`).Scan(&userB) //nolint:errcheck

	_, _ = repo.Create(ctx, &model.Expense{UserID: userA, Amount: 1000, Category: "food", Date: time.Now()})
	_, _ = repo.Create(ctx, &model.Expense{UserID: userB, Amount: 2000, Category: "food", Date: time.Now()})

	listA, _ := repo.List(ctx, model.ListExpensesFilter{UserID: userA})
	listB, _ := repo.List(ctx, model.ListExpensesFilter{UserID: userB})

	if len(listA) != 1 || listA[0].Amount != 1000 {
		t.Errorf("user A should only see their own expense, got %+v", listA)
	}
	if len(listB) != 1 || listB[0].Amount != 2000 {
		t.Errorf("user B should only see their own expense, got %+v", listB)
	}
}

func TestIdempotencyRepo_SaveAndGet(t *testing.T) {
	pool, cleanup := startDB(t)
	defer cleanup()

	repo := repository.NewIdempotencyRepository(pool)
	ctx := context.Background()

	rec := &model.IdempotencyRecord{
		Key: "key-1", RequestHash: "hash-abc",
		Response: []byte(`{"id":"test"}`),
	}
	if err := repo.Save(ctx, rec); err != nil {
		t.Fatalf("save: %v", err)
	}

	got, err := repo.Get(ctx, "key-1")
	if err != nil || got == nil {
		t.Fatalf("get: err=%v, got=%v", err, got)
	}
	if got.RequestHash != "hash-abc" {
		t.Errorf("hash mismatch: %s", got.RequestHash)
	}
}

func TestIdempotencyRepo_GetMissing(t *testing.T) {
	pool, cleanup := startDB(t)
	defer cleanup()

	got, err := repository.NewIdempotencyRepository(pool).Get(context.Background(), "no-such-key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != nil {
		t.Errorf("expected nil, got: %+v", got)
	}
}

func TestIdempotencyRepo_DuplicateSaveIsNoop(t *testing.T) {
	pool, cleanup := startDB(t)
	defer cleanup()

	repo := repository.NewIdempotencyRepository(pool)
	ctx := context.Background()

	r1 := &model.IdempotencyRecord{Key: "dup", RequestHash: "hash1", Response: []byte(`{}`)}
	r2 := &model.IdempotencyRecord{Key: "dup", RequestHash: "hash2", Response: []byte(`{"different":true}`)}

	_ = repo.Save(ctx, r1)
	if err := repo.Save(ctx, r2); err != nil {
		t.Fatalf("second save should be noop, got: %v", err)
	}

	got, _ := repo.Get(ctx, "dup")
	if got.RequestHash != "hash1" {
		t.Errorf("original hash should be preserved, got %s", got.RequestHash)
	}
}
