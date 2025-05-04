'use client';

import { useEffect } from 'react';
import { useRouter, usePathname } from 'next/navigation';
import { useAuth as useAuthContext } from '@/contexts/AuthContext';

export function useAuth(requireAdmin = true) {
  const { isAuthenticated, isAdmin, isLoading, user } = useAuthContext();
  const router = useRouter();
  const pathname = usePathname();

  useEffect(() => {
    // Skip during server-side rendering or while loading
    if (typeof window === 'undefined' || isLoading) {
      return;
    }

    const isAuthPage = pathname === '/signin' || pathname === '/signup';

    // If on auth page and already authenticated, redirect to dashboard
    if (isAuthPage && isAuthenticated) {
      router.push('/');
      return;
    }

    // If not authenticated and not on auth page, redirect to login
    if (!isAuthenticated && !isAuthPage) {
      router.push(`/signin?callbackUrl=${encodeURIComponent(pathname)}`);
      return;
    }

    // If authenticated but not admin and admin is required, redirect to unauthorized page
    if (isAuthenticated && requireAdmin && !isAdmin) {
      // You could create a dedicated unauthorized page
      router.push('/signin');
      return;
    }
  }, [isAuthenticated, isAdmin, isLoading, pathname, requireAdmin, router, user]);

  return {
    isAuthenticated,
    isAdmin,
    isLoading,
    user
  };
}
