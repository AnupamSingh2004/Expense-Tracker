package handler

import (
	"encoding/json"
	"net/http"

	"github.com/fenmo/expense-tracker/internal/middleware"
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
		writeJSON(w, http.StatusBadRequest, errResponse{Code: "INVALID_JSON", Message: "request body must be valid JSON"})
		return
	}
	input.UserID = middleware.UserIDFromCtx(r.Context())

	idempKey := r.Header.Get("Idempotency-Key")
	expense, cached, err := h.svc.CreateExpense(r.Context(), idempKey, input)
	if err != nil {
		writeError(w, h.logger, err)
		return
	}

	status := http.StatusCreated
	if cached {
		status = http.StatusOK
	}
	writeJSON(w, status, expense)
}

func (h *ExpenseHandler) List(w http.ResponseWriter, r *http.Request) {
	filter := model.ListExpensesFilter{
		UserID:   middleware.UserIDFromCtx(r.Context()),
		Category: r.URL.Query().Get("category"),
		SortBy:   r.URL.Query().Get("sort"),
	}
	expenses, err := h.svc.ListExpenses(r.Context(), filter)
	if err != nil {
		writeError(w, h.logger, err)
		return
	}
	if expenses == nil {
		expenses = []*model.Expense{}
	}
	writeJSON(w, http.StatusOK, map[string]any{"expenses": expenses})
}
