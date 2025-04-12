'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import { useAuth } from '@/contexts/AuthContext';

interface SignOutButtonProps {
  className?: string;
  children?: React.ReactNode;
}

export default function SignOutButton({ className, children }: SignOutButtonProps) {
  const [isLoading, setIsLoading] = useState(false);
  const router = useRouter();
  const { logout } = useAuth();

  const handleSignOut = async () => {
    try {
      setIsLoading(true);
      await logout();
      router.push('/signin');
    } catch (error) {
      console.error('Error signing out:', error);
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <button
      onClick={handleSignOut}
      disabled={isLoading}
      className={className || 'px-4 py-2 text-white bg-red-600 rounded hover:bg-red-700 disabled:opacity-50'}
    >
      {isLoading ? 'Signing out...' : children || 'Sign Out'}
    </button>
  );
}
