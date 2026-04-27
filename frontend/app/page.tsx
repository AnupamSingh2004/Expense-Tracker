'use client';

import { ExpenseForm } from '@/components/ExpenseForm';
import { ExpenseList } from '@/components/ExpenseList';
import { AuthGuard, LogoutButton } from '@/components/AuthGuard';

export default function Home() {
  return (
    <AuthGuard>
      <main className="min-h-screen bg-slate-50 text-slate-900 font-sans">
        <header className="bg-white border-b border-slate-200 shadow-sm sticky top-0 z-50 backdrop-blur-sm bg-white/80">
          <div className="max-w-6xl mx-auto px-4 py-4 flex items-center justify-between">
            <div>
              <h1 className="text-3xl font-extrabold bg-gradient-to-r from-indigo-600 to-violet-600 bg-clip-text text-transparent">
                Expense Tracker
              </h1>
              <p className="text-sm font-medium text-slate-500 mt-1">Track your personal expenses seamlessly</p>
            </div>
            <LogoutButton />
          </div>
        </header>
        <div className="max-w-6xl mx-auto px-4 py-10 grid gap-8 lg:grid-cols-[400px_1fr] items-start">
          <div className="lg:sticky lg:top-24">
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
