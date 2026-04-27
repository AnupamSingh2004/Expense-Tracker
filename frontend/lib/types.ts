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

export interface RegisterInput {
  email: string;
  password: string;
}

export interface LoginInput {
  email: string;
  password: string;
}

export interface AuthUser {
  id: string;
  email: string;
  created_at: string;
}

export interface AuthResponse {
  token: string;
  user: AuthUser;
}
