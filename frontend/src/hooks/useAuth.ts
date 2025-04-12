import { useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { useAuth as useAuthContext } from '@/contexts/AuthContext';

export function useAuth(requireAuth = true, requiredRole?: string) {
  const { isAuthenticated, user, isLoading } = useAuthContext();
  const router = useRouter();

  useEffect(() => {
    if (!isLoading && requireAuth) {
      if (!isAuthenticated) {
        router.push(`/signin?callbackUrl=${encodeURIComponent(window.location.href)}`);
        return;
      }

      if (requiredRole && user?.role !== requiredRole) {
        router.push('/403');
        return;
      }
    }
  }, [isAuthenticated, isLoading, requireAuth, requiredRole, router, user]);

  return {
    isAuthenticated,
    isLoading,
    user,
    role: user?.role
  };
}
