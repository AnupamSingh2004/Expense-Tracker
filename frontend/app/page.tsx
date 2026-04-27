'use client';

import { ExpenseForm } from '@/components/ExpenseForm';
import { ExpenseList } from '@/components/ExpenseList';
import { AuthGuard, LogoutButton } from '@/components/AuthGuard';

export default function Home() {
  return (
    <AuthGuard>
      <main className="min-h-screen bg-gray-50">
        <header className="bg-white border-b shadow-sm">
          <div className="max-w-5xl mx-auto px-4 py-4 flex items-center justify-between">
            <div>
              <h1 className="text-2xl font-bold text-indigo-700">Expense Tracker</h1>
              <p className="text-sm text-gray-500">Track your personal expenses</p>
            </div>
            <LogoutButton />
          </div>
        </header>
        <div className="max-w-5xl mx-auto px-4 py-8 grid gap-8 lg:grid-cols-[380px_1fr]">
          <div className="lg:sticky lg:top-8 lg:self-start">
            <ExpenseForm />
          </div>
          <div>
            <ExpenseList />
          </div>
        </div>
      </main>
    </AuthGuard>
  );
}
