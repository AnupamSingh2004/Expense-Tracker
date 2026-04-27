package model

import (
	"time"

	"github.com/google/uuid"
)

// Expense stores amount in paise (1 INR = 100 paise) to avoid float arithmetic.
type Expense struct {
	ID          uuid.UUID `json:"id"`
	Amount      int64     `json:"amount"` // paise
	Category    string    `json:"category"`
	Description string    `json:"description"`
	Date        time.Time `json:"date"`
	CreatedAt   time.Time `json:"created_at"`
}

type CreateExpenseInput struct {
	Amount      int64  `json:"amount"` // paise
	Category    string `json:"category"`
	Description string `json:"description"`
	Date        string `json:"date"` // YYYY-MM-DD
}

type ListExpensesFilter struct {
	Category string
}
