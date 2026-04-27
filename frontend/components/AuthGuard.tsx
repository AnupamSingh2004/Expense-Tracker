'use client';

import { useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { isAuthenticated, clearToken } from '@/lib/auth';

export function AuthGuard({ children }: { children: React.ReactNode }) {
  const router = useRouter();
  // localStorage is only available on the client; this component is always client-side
  const authed = typeof window !== 'undefined' && isAuthenticated();

  useEffect(() => {
    if (!authed) {
      router.replace('/login');
    }
  }, [authed, router]);

  // Don't render children until we've confirmed authentication client-side
  if (!authed) return null;
  return <>{children}</>;
}

export function LogoutButton() {
  const router = useRouter();
  function handleLogout() {
    clearToken();
    router.push('/login');
  }
  return (
    <button
      onClick={handleLogout}
      className="text-sm text-gray-500 hover:text-red-600 transition-colors"
    >
      Sign out
    </button>
  );
}
