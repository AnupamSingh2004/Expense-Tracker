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

  const { data: expenses = [], isLoading, isError, error, refetch } = useQuery({
    queryKey: ['expenses', category],
    queryFn: () => listExpenses(category || undefined),
  });

  const total = expenses.reduce((sum, e) => sum + e.amount, 0);

  return (
    <div className="bg-white rounded-xl shadow p-6 space-y-4">
      <div className="flex items-center justify-between flex-wrap gap-3">
        <h2 className="text-lg font-semibold text-gray-800">Expenses</h2>
        <FilterBar category={category} onCategoryChange={setCategory} />
      </div>

      {isLoading && <LoadingSpinner />}

      {isError && (
        <ErrorBanner
          message={(error as Error).message}
          onRetry={() => refetch()}
        />
      )}

      {!isLoading && !isError && expenses.length === 0 && (
        <p className="text-center text-gray-400 py-8 text-sm">No expenses found.</p>
      )}

      {expenses.length > 0 && (
        <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b text-left text-gray-500">
                <th className="pb-2 font-medium">Date</th>
                <th className="pb-2 font-medium">Category</th>
                <th className="pb-2 font-medium">Description</th>
                <th className="pb-2 font-medium text-right">Amount</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-100">
              {expenses.map((e) => (
                <tr key={e.id} className="hover:bg-gray-50 transition-colors">
                  <td className="py-2.5 text-gray-600 whitespace-nowrap">{e.date}</td>
                  <td className="py-2.5">
                    <span className="inline-block bg-indigo-100 text-indigo-700 text-xs px-2 py-0.5 rounded-full capitalize">
                      {e.category}
                    </span>
                  </td>
                  <td className="py-2.5 text-gray-600 max-w-[200px] truncate">
                    {e.description || '—'}
                  </td>
                  <td className="py-2.5 text-right font-medium text-gray-800 whitespace-nowrap">
                    {formatPaise(e.amount)}
                  </td>
                </tr>
              ))}
            </tbody>
            <tfoot>
              <tr className="border-t-2 border-gray-200 font-semibold">
                <td colSpan={3} className="pt-3 text-gray-700">
                  Total ({expenses.length} item{expenses.length !== 1 ? 's' : ''})
                </td>
                <td className="pt-3 text-right text-indigo-700">{formatPaise(total)}</td>
              </tr>
            </tfoot>
          </table>
        </div>
      )}
    </div>
  );
}
