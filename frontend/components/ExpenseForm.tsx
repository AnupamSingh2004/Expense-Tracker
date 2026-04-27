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
    <form onSubmit={handleSubmit} className="bg-white rounded-2xl shadow-xl shadow-slate-200/50 border border-slate-100 p-8 space-y-6">
      <div className="border-b border-slate-100 pb-4">
        <h2 className="text-2xl font-bold tracking-tight text-slate-800">Add Expense</h2>
        <p className="text-sm text-slate-500 mt-1">Record a new transaction</p>
      </div>

      {errorMsg && (
        <p className="text-sm text-red-600 bg-red-50 border border-red-200 rounded-lg p-3 shadow-sm font-medium">
          {errorMsg}
        </p>
      )}

      <div className="grid grid-cols-2 gap-5">
        <div>
          <label htmlFor="amount" className="block text-sm font-semibold text-slate-700 mb-1.5">
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
            className="w-full border border-slate-300 rounded-lg px-4 py-3 text-sm bg-slate-50 focus:bg-white focus:outline-none focus:ring-4 focus:ring-indigo-500/20 focus:border-indigo-500 transition-all duration-200 placeholder:text-slate-400 font-medium"
          />
        </div>
        <div>
          <label htmlFor="category" className="block text-sm font-semibold text-slate-700 mb-1.5">
            Category
          </label>
          <select
            id="category"
            name="category"
            value={form.category}
            onChange={handleChange}
            className="w-full border border-slate-300 rounded-lg px-4 py-3 text-sm bg-slate-50 focus:bg-white focus:outline-none focus:ring-4 focus:ring-indigo-500/20 focus:border-indigo-500 transition-all duration-200 font-medium text-slate-800"
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
        <label htmlFor="date" className="block text-sm font-semibold text-slate-700 mb-1.5">
          Date
        </label>
        <input
          id="date"
          name="date"
          type="date"
          value={form.date}
          onChange={handleChange}
          required
          className="w-full border border-slate-300 rounded-lg px-4 py-3 text-sm bg-slate-50 focus:bg-white focus:outline-none focus:ring-4 focus:ring-indigo-500/20 focus:border-indigo-500 transition-all duration-200 font-medium text-slate-800"
        />
      </div>

      <div>
        <label htmlFor="description" className="block text-sm font-semibold text-slate-700 mb-1.5">
          Description
        </label>
        <textarea
          id="description"
          name="description"
          value={form.description}
          onChange={handleChange}
          rows={2}
          placeholder="Optional notes..."
          className="w-full border border-slate-300 rounded-lg px-4 py-3 text-sm bg-slate-50 focus:bg-white focus:outline-none focus:ring-4 focus:ring-indigo-500/20 focus:border-indigo-500 transition-all duration-200 resize-none placeholder:text-slate-400 font-medium text-slate-800"
        />
      </div>

      <button
        type="submit"
        disabled={isSubmitting}
        className="w-full mt-2 bg-gradient-to-r from-indigo-600 to-violet-600 hover:from-indigo-700 hover:to-violet-700 shadow-md shadow-indigo-500/30 disabled:opacity-50 disabled:cursor-not-allowed text-white font-bold py-3.5 px-4 rounded-xl transition-all duration-200 transform hover:-translate-y-0.5 active:translate-y-0"
      >
        {isSubmitting ? 'Saving…' : 'Add Expense'}
      </button>
    </form>
  );
}
