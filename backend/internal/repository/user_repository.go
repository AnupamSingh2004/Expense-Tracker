package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/fenmo/expense-tracker/internal/model"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository interface {
	Create(ctx context.Context, email, passwordHash string) (*model.User, error)
	GetByEmail(ctx context.Context, email string) (*model.User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*model.User, error)
}

type pgUserRepo struct{ pool *pgxpool.Pool }

func NewUserRepository(pool *pgxpool.Pool) UserRepository {
	return &pgUserRepo{pool: pool}
}

func (r *pgUserRepo) Create(ctx context.Context, email, passwordHash string) (*model.User, error) {
	const q = `
		INSERT INTO users (email, password_hash)
		VALUES ($1, $2)
		RETURNING id, email, created_at`
	u := &model.User{}
	err := r.pool.QueryRow(ctx, q, email, passwordHash).Scan(&u.ID, &u.Email, &u.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("user create: %w", err)
	}
	return u, nil
}

func (r *pgUserRepo) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	const q = `SELECT id, email, password_hash, created_at FROM users WHERE email = $1`
	u := &model.User{}
	err := r.pool.QueryRow(ctx, q, email).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("user get by email: %w", err)
	}
	return u, nil
}

func (r *pgUserRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	const q = `SELECT id, email, created_at FROM users WHERE id = $1`
	u := &model.User{}
	err := r.pool.QueryRow(ctx, q, id).Scan(&u.ID, &u.Email, &u.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("user get by id: %w", err)
	}
	return u, nil
}
