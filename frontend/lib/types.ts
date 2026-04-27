export interface Expense {
  id: string;
  amount: number; // paise
  category: string;
  description: string;
  date: string;
  created_at: string;
}

export interface CreateExpensePayload {
  amount: number; // paise
  category: string;
  description: string;
  date: string;
}

export interface ListExpensesResponse {
  expenses: Expense[];
}

export interface ApiError {
  code: string;
  message: string;
}
