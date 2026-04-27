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

type pgExpenseRepo struct{ pool *pgxpool.Pool }

func NewExpenseRepository(pool *pgxpool.Pool) ExpenseRepository {
	return &pgExpenseRepo{pool: pool}
}

func (r *pgExpenseRepo) Create(ctx context.Context, e *model.Expense) (*model.Expense, error) {
	e.ID = uuid.New()
	e.CreatedAt = time.Now().UTC()

	const q = `
		INSERT INTO expenses (id, amount, category, description, date, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, amount, category, description, date, created_at`

	out := &model.Expense{}
	err := r.pool.QueryRow(ctx, q, e.ID, e.Amount, e.Category, e.Description, e.Date, e.CreatedAt).
		Scan(&out.ID, &out.Amount, &out.Category, &out.Description, &out.Date, &out.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("expense create: %w", err)
	}
	return out, nil
}

func (r *pgExpenseRepo) List(ctx context.Context, f model.ListExpensesFilter) ([]*model.Expense, error) {
	var (
		args  []any
		where string
	)
	if f.Category != "" {
		args = append(args, f.Category)
		where = "WHERE category = $1"
	}

	q := fmt.Sprintf(`
		SELECT id, amount, category, description, date, created_at
		FROM expenses %s
		ORDER BY date DESC, created_at DESC`, where)

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
