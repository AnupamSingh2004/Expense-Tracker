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

// --- in-memory mocks ---

type mockExpRepo struct{ items []*model.Expense }

func (m *mockExpRepo) Create(_ context.Context, e *model.Expense) (*model.Expense, error) {
	m.items = append(m.items, e)
	return e, nil
}
func (m *mockExpRepo) List(_ context.Context, _ model.ListExpensesFilter) ([]*model.Expense, error) {
	return m.items, nil
}

type mockIdempRepo struct {
	store map[string]*model.IdempotencyRecord
}

func newMockIdemp() *mockIdempRepo {
	return &mockIdempRepo{store: map[string]*model.IdempotencyRecord{}}
}
func (m *mockIdempRepo) Get(_ context.Context, key string) (*model.IdempotencyRecord, error) {
	return m.store[key], nil
}
func (m *mockIdempRepo) Save(_ context.Context, rec *model.IdempotencyRecord) error {
	m.store[rec.Key] = rec
	return nil
}

var _ repository.ExpenseRepository = &mockExpRepo{}
var _ repository.IdempotencyRepository = &mockIdempRepo{}

func newSvc(exp *mockExpRepo, idemp *mockIdempRepo) service.ExpenseService {
	return service.NewExpenseService(exp, idemp, zap.NewNop())
}

func validInput() model.CreateExpenseInput {
	return model.CreateExpenseInput{Amount: 5000, Category: "food", Description: "lunch", Date: "2024-01-15"}
}

// --- tests ---

func TestCreate_Success(t *testing.T) {
	svc := newSvc(&mockExpRepo{}, newMockIdemp())
	e, cached, err := svc.CreateExpense(context.Background(), "", validInput())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cached {
		t.Error("expected cached=false")
	}
	if e.Amount != 5000 {
		t.Errorf("got amount %d, want 5000", e.Amount)
	}
}

func TestCreate_IdempotencyHit(t *testing.T) {
	expRepo := &mockExpRepo{}
	svc := newSvc(expRepo, newMockIdemp())

	input := validInput()
	_, _, _ = svc.CreateExpense(context.Background(), "key-1", input)
	_, cached, err := svc.CreateExpense(context.Background(), "key-1", input)
	if err != nil {
		t.Fatalf("second call error: %v", err)
	}
	if !cached {
		t.Error("expected cached=true on replay")
	}
	if len(expRepo.items) != 1 {
		t.Errorf("expected 1 DB write, got %d", len(expRepo.items))
	}
}

func TestCreate_IdempotencyConflict(t *testing.T) {
	svc := newSvc(&mockExpRepo{}, newMockIdemp())
	_, _, _ = svc.CreateExpense(context.Background(), "key-1", validInput())

	diff := validInput()
	diff.Amount = 9999
	_, _, err := svc.CreateExpense(context.Background(), "key-1", diff)
	if err == nil {
		t.Fatal("expected conflict error")
	}
}

func TestCreate_InvalidAmount(t *testing.T) {
	svc := newSvc(&mockExpRepo{}, newMockIdemp())
	input := validInput()
	input.Amount = 0
	_, _, err := svc.CreateExpense(context.Background(), "", input)
	if err == nil {
		t.Fatal("expected validation error for zero amount")
	}
}

func TestCreate_InvalidDate(t *testing.T) {
	svc := newSvc(&mockExpRepo{}, newMockIdemp())
	input := validInput()
	input.Date = "not-a-date"
	_, _, err := svc.CreateExpense(context.Background(), "", input)
	if err == nil {
		t.Fatal("expected error for invalid date")
	}
}

func TestCreate_MissingCategory(t *testing.T) {
	svc := newSvc(&mockExpRepo{}, newMockIdemp())
	input := validInput()
	input.Category = ""
	_, _, err := svc.CreateExpense(context.Background(), "", input)
	if err == nil {
		t.Fatal("expected error for missing category")
	}
}

func TestIdempotencyCache_RoundTrip(t *testing.T) {
	svc := newSvc(&mockExpRepo{}, newMockIdemp())
	input := validInput()
	orig, _, _ := svc.CreateExpense(context.Background(), "rt-key", input)
	cached, _, err := svc.CreateExpense(context.Background(), "rt-key", input)
	if err != nil {
		t.Fatalf("cache retrieval: %v", err)
	}
	a, _ := json.Marshal(orig)
	b, _ := json.Marshal(cached)
	if string(a) != string(b) {
		t.Errorf("round-trip mismatch:\norig:   %s\ncached: %s", a, b)
	}
}

func TestList_Empty(t *testing.T) {
	svc := newSvc(&mockExpRepo{}, newMockIdemp())
	list, err := svc.ListExpenses(context.Background(), model.ListExpensesFilter{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(list) != 0 {
		t.Errorf("expected 0, got %d", len(list))
	}
}
