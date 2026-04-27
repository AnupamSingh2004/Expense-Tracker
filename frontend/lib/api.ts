import { CreateExpensePayload, Expense, ListExpensesResponse, ApiError, AuthResponse, LoginInput, RegisterInput } from './types';
import { getToken, clearToken } from './auth';

const BASE_URL = process.env.NEXT_PUBLIC_API_URL ?? 'http://localhost:8080';

function authHeaders(): Record<string, string> {
  const token = getToken();
  return token ? { Authorization: `Bearer ${token}` } : {};
}

async function handleResponse<T>(res: Response): Promise<T> {
  if (res.status === 401) {
    clearToken();
    window.location.href = '/login';
    throw new Error('Session expired. Please log in again.');
  }
  if (!res.ok) {
    const err: ApiError = await res.json();
    throw new Error(err.message ?? `HTTP ${res.status}`);
  }
  return res.json() as Promise<T>;
}

export async function register(input: RegisterInput): Promise<AuthResponse> {
  const res = await fetch(`${BASE_URL}/auth/register`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(input),
  });
  return handleResponse<AuthResponse>(res);
}

export async function login(input: LoginInput): Promise<AuthResponse> {
  const res = await fetch(`${BASE_URL}/auth/login`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(input),
  });
  return handleResponse<AuthResponse>(res);
}

export async function createExpense(
  payload: CreateExpensePayload,
  idempotencyKey: string
): Promise<Expense> {
  const res = await fetch(`${BASE_URL}/expenses`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'Idempotency-Key': idempotencyKey,
      ...authHeaders(),
    },
    body: JSON.stringify(payload),
  });
  return handleResponse<Expense>(res);
}

export async function listExpenses(category?: string, sort = 'date_desc'): Promise<Expense[]> {
  const params = new URLSearchParams({ sort });
  if (category) params.set('category', category);

  const res = await fetch(`${BASE_URL}/expenses?${params.toString()}`, {
    cache: 'no-store',
    headers: authHeaders(),
  });
  const data = await handleResponse<ListExpensesResponse>(res);
  return data.expenses ?? [];
}
