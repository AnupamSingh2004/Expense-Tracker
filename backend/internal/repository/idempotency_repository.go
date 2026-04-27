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

type pgIdempRepo struct{ pool *pgxpool.Pool }

func NewIdempotencyRepository(pool *pgxpool.Pool) IdempotencyRepository {
	return &pgIdempRepo{pool: pool}
}

func (r *pgIdempRepo) Get(ctx context.Context, key string) (*model.IdempotencyRecord, error) {
	const q = `SELECT key, request_hash, response, created_at FROM idempotency_keys WHERE key = $1`
	rec := &model.IdempotencyRecord{}
	var resp []byte
	err := r.pool.QueryRow(ctx, q, key).Scan(&rec.Key, &rec.RequestHash, &resp, &rec.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("idempotency get: %w", err)
	}
	rec.Response = json.RawMessage(resp)
	return rec, nil
}

func (r *pgIdempRepo) Save(ctx context.Context, rec *model.IdempotencyRecord) error {
	const q = `
		INSERT INTO idempotency_keys (key, request_hash, response, created_at)
		VALUES ($1, $2, $3, NOW())
		ON CONFLICT (key) DO NOTHING`
	_, err := r.pool.Exec(ctx, q, rec.Key, rec.RequestHash, []byte(rec.Response))
	if err != nil {
		return fmt.Errorf("idempotency save: %w", err)
	}
	return nil
}
