import { render, screen } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { ExpenseList } from '@/components/ExpenseList';
import * as api from '@/lib/api';
import type { Expense } from '@/lib/types';

jest.mock('@/lib/api');

function Wrapper({ children }: { children: React.ReactNode }) {
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return <QueryClientProvider client={qc}>{children}</QueryClientProvider>;
}

const mockExpenses: Expense[] = [
  { id: '1', amount: 5000, category: 'food', description: 'lunch', date: '2024-01-15', created_at: '' },
  { id: '2', amount: 10000, category: 'travel', description: 'uber', date: '2024-01-14', created_at: '' },
];

describe('ExpenseList', () => {
  it('renders expenses and computes total', async () => {
    (api.listExpenses as jest.Mock).mockResolvedValue(mockExpenses);
    render(<ExpenseList />, { wrapper: Wrapper });
    expect(await screen.findByText('lunch')).toBeInTheDocument();
    expect(screen.getByText('uber')).toBeInTheDocument();
    // 5000 + 10000 paise = ₹150
    expect(screen.getByText(/₹150/)).toBeInTheDocument();
  });

  it('shows empty state when no expenses', async () => {
    (api.listExpenses as jest.Mock).mockResolvedValue([]);
    render(<ExpenseList />, { wrapper: Wrapper });
    expect(await screen.findByText(/no expenses found/i)).toBeInTheDocument();
  });

  it('shows error banner on API failure', async () => {
    (api.listExpenses as jest.Mock).mockRejectedValue(new Error('Network error'));
    render(<ExpenseList />, { wrapper: Wrapper });
    expect(await screen.findByText(/network error/i)).toBeInTheDocument();
  });
});
