'use client';

import { createContext, useContext, ReactNode } from 'react';
import { useSession } from 'next-auth/react';

interface AuthContextType {
  isAuthenticated: boolean;
  accessToken: string | null;
  userRole: string | null;
}

const AuthContext = createContext<AuthContextType>({
  isAuthenticated: false,
  accessToken: null,
  userRole: null,
});

export function AuthProvider({ children }: { children: ReactNode }) {
  const { data: session } = useSession();

  const value = {
    isAuthenticated: !!session,
    accessToken: session?.accessToken || null,
    userRole: session?.user?.role || null,
  };

  return (
    <AuthContext.Provider value={value}>
      {children}
    </AuthContext.Provider>
  );
}

export const useAuth = () => useContext(AuthContext);