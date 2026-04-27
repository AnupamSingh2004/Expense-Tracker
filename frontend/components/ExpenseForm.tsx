'use client';

import { useRef, useState } from 'react';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { createExpense } from '@/lib/api';
import { generateIdempotencyKey } from '@/lib/idempotency';
import { rupeesToPaise } from '@/lib/money';

const CATEGORIES = ['food', 'travel', 'utilities', 'entertainment', 'health', 'other'];

interface FormState {
  amount: string;
  category: string;
  description: string;
  date: string;
}

const defaultForm = (): FormState => ({
  amount: '',
  category: 'food',
  description: '',
  date: new Date().toISOString().split('T')[0] ?? '',
});

export function ExpenseForm() {
  const [form, setForm] = useState<FormState>(defaultForm);
  const [validationError, setValidationError] = useState<string | null>(null);
  // Stable per submission — rotated on success to prevent accidental replay
  const idempKey = useRef(generateIdempotencyKey());

  const qc = useQueryClient();
  const mutation = useMutation({
    mutationFn: () =>
      createExpense(
        {
          amount: rupeesToPaise(form.amount),
          category: form.category,
          description: form.description,
          date: form.date,
        },
        idempKey.current
      ),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['expenses'] });
      setForm(defaultForm());
      idempKey.current = generateIdempotencyKey();
      setValidationError(null);
    },
  });

  function handleChange(
    e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement | HTMLTextAreaElement>
  ) {
    setForm((f) => ({ ...f, [e.target.name]: e.target.value }));
  }

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setValidationError(null);
    const amount = parseFloat(form.amount);
    if (isNaN(amount) || amount <= 0) {
      setValidationError('Amount must be a positive number');
      return;
    }
    if (!form.date) {
      setValidationError('Date is required');
      return;
    }
    mutation.mutate();
  }

  const isSubmitting = mutation.isPending;
  const errorMsg = validationError ?? (mutation.error as Error | null)?.message ?? null;

  return (
    <form onSubmit={handleSubmit} className="bg-white rounded-xl shadow p-6 space-y-4">
      <h2 className="text-lg font-semibold text-gray-800">Add Expense</h2>

      {errorMsg && (
        <p className="text-sm text-red-600 bg-red-50 border border-red-200 rounded p-2">
          {errorMsg}
        </p>
      )}

      <div className="grid grid-cols-2 gap-4">
        <div>
          <label htmlFor="amount" className="block text-sm font-medium text-gray-700 mb-1">
            Amount (₹)
          </label>
          <input
            id="amount"
            name="amount"
            type="number"
            step="0.01"
            min="0.01"
            value={form.amount}
            onChange={handleChange}
            required
            placeholder="0.00"
            className="w-full border rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
          />
        </div>
        <div>
          <label htmlFor="category" className="block text-sm font-medium text-gray-700 mb-1">
            Category
          </label>
          <select
            id="category"
            name="category"
            value={form.category}
            onChange={handleChange}
            className="w-full border rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
          >
            {CATEGORIES.map((c) => (
              <option key={c} value={c}>
                {c.charAt(0).toUpperCase() + c.slice(1)}
              </option>
            ))}
          </select>
        </div>
      </div>

      <div>
        <label htmlFor="date" className="block text-sm font-medium text-gray-700 mb-1">
          Date
        </label>
        <input
          id="date"
          name="date"
          type="date"
          value={form.date}
          onChange={handleChange}
          required
          className="w-full border rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
        />
      </div>

      <div>
        <label htmlFor="description" className="block text-sm font-medium text-gray-700 mb-1">
          Description
        </label>
        <textarea
          id="description"
          name="description"
          value={form.description}
          onChange={handleChange}
          rows={2}
          placeholder="Optional notes..."
          className="w-full border rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500 resize-none"
        />
      </div>

      <button
        type="submit"
        disabled={isSubmitting}
        className="w-full bg-indigo-600 hover:bg-indigo-700 disabled:bg-indigo-300 text-white font-medium py-2 px-4 rounded-md transition-colors"
      >
        {isSubmitting ? 'Saving…' : 'Add Expense'}
      </button>
    </form>
  );
}
