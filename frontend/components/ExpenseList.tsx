'use client';

import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { listExpenses } from '@/lib/api';
import { formatPaise } from '@/lib/money';
import { FilterBar } from './FilterBar';
import { LoadingSpinner } from './LoadingSpinner';
import { ErrorBanner } from './ErrorBanner';

export function ExpenseList() {
  const [category, setCategory] = useState('');
  const [sort, setSort] = useState<'date_desc' | 'date_asc'>('date_desc');

  const { data: expenses = [], isLoading, isError, error, refetch } = useQuery({
    queryKey: ['expenses', category, sort],
    queryFn: () => listExpenses(category || undefined, sort),
  });

  const total = expenses.reduce((sum, e) => sum + e.amount, 0);

  return (
    <div className="bg-white rounded-2xl shadow-xl shadow-slate-200/50 border border-slate-100 p-8 space-y-6">
      <div className="flex items-center justify-between flex-wrap gap-4 border-b border-slate-100 pb-4">
        <div>
          <h2 className="text-2xl font-bold tracking-tight text-slate-800">Expenses</h2>
          <p className="text-sm text-slate-500 mt-1">Your recent transactions</p>
        </div>
        <div className="flex items-center gap-3 flex-wrap">
          <FilterBar category={category} onCategoryChange={setCategory} />
          <button
            onClick={() => setSort((s) => (s === 'date_desc' ? 'date_asc' : 'date_desc'))}
            className="flex items-center gap-2 text-sm font-medium border border-slate-300 rounded-lg px-4 py-2 hover:bg-slate-50 hover:border-slate-400 focus:ring-4 focus:ring-indigo-500/20 transition-all duration-200 text-slate-700 shadow-sm"
          >
            Date {sort === 'date_desc' ? '↓' : '↑'}
          </button>
        </div>
      </div>

      {isLoading && <LoadingSpinner />}

      {isError && (
        <ErrorBanner
          message={(error as Error).message}
          onRetry={() => refetch()}
        />
      )}

      {!isLoading && !isError && expenses.length === 0 && (
        <div className="text-center py-12 px-4 border-2 border-dashed border-slate-200 rounded-xl bg-slate-50">
          <p className="text-slate-500 font-medium">No expenses found.</p>
          <p className="text-sm text-slate-400 mt-1">Add a new expense to see it here.</p>
        </div>
      )}

      {expenses.length > 0 && (
        <div className="overflow-x-auto rounded-xl border border-slate-100 bg-white">
          <table className="w-full text-sm">
            <thead className="bg-slate-50">
              <tr className="border-b border-slate-200 text-left text-slate-600">
                <th className="py-3 px-5 font-semibold">Date</th>
                <th className="py-3 px-5 font-semibold">Category</th>
                <th className="py-3 px-5 font-semibold">Description</th>
                <th className="py-3 px-5 font-semibold text-right">Amount</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-100">
              {expenses.map((e) => (
                <tr key={e.id} className="hover:bg-indigo-50/40 transition-colors duration-150 group">
                  <td className="py-4 px-5 text-slate-600 whitespace-nowrap font-medium">
                    {new Date(e.date).toLocaleDateString('en-IN', {
                      year: 'numeric',
                      month: 'short',
                      day: 'numeric',
                      timeZone: 'UTC',
                    })}
                  </td>
                  <td className="py-4 px-5">
                    <span className="inline-flex items-center bg-indigo-50 text-indigo-700 border border-indigo-100 shadow-sm text-xs font-bold px-3 py-1 rounded-full capitalize tracking-wide">
                      {e.category}
                    </span>
                  </td>
                  <td className="py-4 px-5 text-slate-600 max-w-[240px] truncate group-hover:text-slate-900 transition-colors">
                    {e.description || '—'}
                  </td>
                  <td className="py-4 px-5 text-right font-bold text-slate-800 whitespace-nowrap text-base">
                    {formatPaise(e.amount)}
                  </td>
                </tr>
              ))}
            </tbody>
            <tfoot className="bg-slate-50">
              <tr className="border-t border-slate-200 font-semibold shadow-[0_-4px_6px_-1px_rgba(0,0,0,0.05)] relative z-10">
                <td colSpan={3} className="py-4 px-5 text-slate-700 uppercase tracking-widest text-xs">
                  Total ({expenses.length} item{expenses.length !== 1 ? 's' : ''})
                </td>
                <td className="py-4 px-5 text-right font-extrabold text-transparent bg-clip-text bg-gradient-to-r from-indigo-700 to-violet-700 text-lg">
                  {formatPaise(total)}
                </td>
              </tr>
            </tfoot>
          </table>
        </div>
      )}
    </div>
  );
}
