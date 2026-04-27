import { CreateExpensePayload, Expense, ListExpensesResponse, ApiError } from './types';

const BASE_URL = process.env.NEXT_PUBLIC_API_URL ?? 'http://localhost:8080';

async function handleResponse<T>(res: Response): Promise<T> {
  if (!res.ok) {
    const err: ApiError = await res.json();
    throw new Error(err.message ?? `HTTP ${res.status}`);
  }
  return res.json() as Promise<T>;
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
    },
    body: JSON.stringify(payload),
  });
  return handleResponse<Expense>(res);
}

export async function listExpenses(category?: string): Promise<Expense[]> {
  const params = new URLSearchParams({ sort: 'date_desc' });
  if (category) params.set('category', category);

  const res = await fetch(`${BASE_URL}/expenses?${params.toString()}`, {
    cache: 'no-store',
  });
  const data = await handleResponse<ListExpensesResponse>(res);
  return data.expenses ?? [];
}
