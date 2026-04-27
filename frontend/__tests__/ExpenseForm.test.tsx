import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { ExpenseForm } from '@/components/ExpenseForm';
import * as api from '@/lib/api';

jest.mock('@/lib/api');

function Wrapper({ children }: { children: React.ReactNode }) {
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false }, mutations: { retry: false } } });
  return <QueryClientProvider client={qc}>{children}</QueryClientProvider>;
}

describe('ExpenseForm', () => {
  it('renders all form fields', () => {
    render(<ExpenseForm />, { wrapper: Wrapper });
    expect(screen.getByLabelText(/amount/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/category/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/date/i)).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /add expense/i })).toBeInTheDocument();
  });

  it('shows validation error for zero amount', async () => {
    render(<ExpenseForm />, { wrapper: Wrapper });
    fireEvent.change(screen.getByLabelText(/amount/i), { target: { value: '0' } });
    fireEvent.submit(screen.getByRole('button', { name: /add expense/i }).closest('form')!);
    await waitFor(() =>
      expect(screen.getByText(/amount must be a positive number/i)).toBeInTheDocument()
    );
  });

  it('disables button while submitting', async () => {
    (api.createExpense as jest.Mock).mockImplementation(
      () => new Promise((resolve) => setTimeout(resolve, 500))
    );
    render(<ExpenseForm />, { wrapper: Wrapper });
    fireEvent.change(screen.getByLabelText(/amount/i), { target: { value: '100' } });
    fireEvent.submit(screen.getByRole('button').closest('form')!);
    expect(await screen.findByRole('button', { name: /saving/i })).toBeDisabled();
  });

  it('resets form on successful submission', async () => {
    (api.createExpense as jest.Mock).mockResolvedValue({ id: '1', amount: 10000, category: 'food', description: '', date: '2024-01-15', created_at: '' });
    render(<ExpenseForm />, { wrapper: Wrapper });
    fireEvent.change(screen.getByLabelText(/amount/i), { target: { value: '100' } });
    fireEvent.submit(screen.getByRole('button').closest('form')!);
    await waitFor(() =>
      expect((screen.getByLabelText(/amount/i) as HTMLInputElement).value).toBe('')
    );
  });
});
