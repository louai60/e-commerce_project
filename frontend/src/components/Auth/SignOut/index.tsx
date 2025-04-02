'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import { AuthService } from '@/services/auth.service';

interface SignOutButtonProps {
  className?: string;
  children?: React.ReactNode;
}

export default function SignOutButton({ className, children }: SignOutButtonProps) {
  const [isLoading, setIsLoading] = useState(false);
  const router = useRouter();

  const handleSignOut = async () => {
    try {
      setIsLoading(true);
    //   await AuthService.logout();
      // The redirect will be handled by NextAuth signOut
    } catch (error) {
      console.error('Error signing out:', error);
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