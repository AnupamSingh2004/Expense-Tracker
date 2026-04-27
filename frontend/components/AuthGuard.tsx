'use client';

import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import { isAuthenticated, clearToken } from '@/lib/auth';
import { LoadingSpinner } from './LoadingSpinner';

export function AuthGuard({ children }: { children: React.ReactNode }) {
  const router = useRouter();
  const [checking, setChecking] = useState(true);

  useEffect(() => {
    if (!isAuthenticated()) {
      router.replace('/login');
    } else {
      setChecking(false);
    }
  }, [router]);

  if (checking) return <LoadingSpinner />;
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
