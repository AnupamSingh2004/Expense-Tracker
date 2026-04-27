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
	CreateExpense(ctx context.Context, idempKey string, input model.CreateExpenseInput) (*model.Expense, bool, error)
	ListExpenses(ctx context.Context, filter model.ListExpensesFilter) ([]*model.Expense, error)
}

type expenseService struct {
	expRepo   repository.ExpenseRepository
	idempRepo repository.IdempotencyRepository
	logger    *zap.Logger
}

func NewExpenseService(
	expRepo repository.ExpenseRepository,
	idempRepo repository.IdempotencyRepository,
	logger *zap.Logger,
) ExpenseService {
	return &expenseService{expRepo: expRepo, idempRepo: idempRepo, logger: logger}
}

func (s *expenseService) CreateExpense(
	ctx context.Context, idempKey string, input model.CreateExpenseInput,
) (*model.Expense, bool, error) {
	if err := validate(input); err != nil {
		return nil, false, err
	}

	requestHash := hashInput(input)

	if idempKey != "" {
		rec, err := s.idempRepo.Get(ctx, idempKey)
		if err != nil {
			return nil, false, fmt.Errorf("idempotency lookup: %w", err)
		}
		if rec != nil {
			if rec.RequestHash != requestHash {
				s.logger.Warn("idempotency conflict",
					zap.String("key", idempKey),
					zap.String("stored_hash", rec.RequestHash),
					zap.String("request_hash", requestHash),
				)
				return nil, false, model.ErrDuplicateKey
			}
			var cached model.Expense
			if err := json.Unmarshal(rec.Response, &cached); err != nil {
				return nil, false, fmt.Errorf("decode cached response: %w", err)
			}
			return &cached, true, nil
		}
	}

	date, err := time.Parse("2006-01-02", input.Date)
	if err != nil {
		return nil, false, fmt.Errorf("%w: invalid date, expected YYYY-MM-DD", model.ErrInvalidInput)
	}

	created, err := s.expRepo.Create(ctx, &model.Expense{
		Amount:      input.Amount,
		Category:    input.Category,
		Description: input.Description,
		Date:        date,
	})
	if err != nil {
		return nil, false, fmt.Errorf("create expense: %w", err)
	}

	if idempKey != "" {
		resp, _ := json.Marshal(created)
		if saveErr := s.idempRepo.Save(ctx, &model.IdempotencyRecord{
			Key: idempKey, RequestHash: requestHash, Response: json.RawMessage(resp),
		}); saveErr != nil {
			s.logger.Error("failed to persist idempotency record", zap.Error(saveErr))
		}
	}

	return created, false, nil
}

func (s *expenseService) ListExpenses(ctx context.Context, filter model.ListExpensesFilter) ([]*model.Expense, error) {
	return s.expRepo.List(ctx, filter)
}

func validate(input model.CreateExpenseInput) error {
	if input.Amount <= 0 {
		return fmt.Errorf("%w: amount must be positive paise", model.ErrInvalidInput)
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
