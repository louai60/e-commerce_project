'use client';

import { createContext, useContext, useState, useEffect, ReactNode, useCallback } from 'react';
import { useRouter, usePathname } from 'next/navigation';
import { AuthService } from '@/services/auth.service';

interface User {
  id: string | number;
  email: string;
  username?: string;
  role: string;
  firstName?: string;
  lastName?: string;
}

interface AuthContextType {
  isAuthenticated: boolean;
  user: User | null;
  accessToken: string | null;
  login: (token: string, userData: User) => void;
  logout: () => void;
  isLoading: boolean;
  isAdmin: boolean;
}

const AuthContext = createContext<AuthContextType>({
  isAuthenticated: false,
  user: null,
  accessToken: null,
  login: () => {},
  logout: () => {},
  isLoading: true,
  isAdmin: false,
});

export function AuthProvider({ children }: { children: ReactNode }) {
  const [accessToken, setAccessToken] = useState<string | null>(null);
  const [user, setUser] = useState<User | null>(null);
  const [isAuthenticated, setIsAuthenticated] = useState<boolean>(false);
  const [isLoading, setIsLoading] = useState<boolean>(true);
  const [isAdmin, setIsAdmin] = useState<boolean>(false);
  const router = useRouter();
  const pathname = usePathname();

  // Define handleLogout function with useCallback to prevent dependency issues
  const handleLogout = useCallback(() => {
    AuthService.logout();
    setAccessToken(null);
    setUser(null);
    setIsAuthenticated(false);
    setIsAdmin(false);
    router.push('/signin');
  }, [router]);

  // Initialize auth state from localStorage
  useEffect(() => {
    const initAuth = () => {
      try {
        const token = localStorage.getItem('access_token');
        const userStr = localStorage.getItem('user');

        if (token && userStr) {
          const userData = JSON.parse(userStr);
          setAccessToken(token);
          setUser(userData);
          setIsAuthenticated(true);
          setIsAdmin(userData.role === 'admin');
        } else {
          // If no token or user data, and not on auth page, redirect to login
          const isAuthPage = pathname === '/signin' || pathname === '/signup';
          if (!isAuthPage) {
            router.push('/signin');
          }
        }
      } catch (error) {
        console.error('Error initializing auth state:', error);
        // Clear potentially corrupted data
        localStorage.removeItem('access_token');
        localStorage.removeItem('user');
      } finally {
        setIsLoading(false);
      }
    };

    if (typeof window !== 'undefined') {
      initAuth();
    } else {
      setIsLoading(false);
    }
  }, [pathname, router]);

  // Listen for storage events (for multi-tab support)
  useEffect(() => {
    const handleStorageChange = (e: StorageEvent) => {
      if (e.key === 'access_token' && !e.newValue) {
        handleLogout();
      } else if (e.key === 'access_token' && e.newValue !== accessToken) {
        // Token was updated
        setAccessToken(e.newValue);
        setIsAuthenticated(true);
      } else if (e.key === 'user') {
        if (!e.newValue) {
          setUser(null);
          setIsAdmin(false);
        } else {
          try {
            const userData = JSON.parse(e.newValue);
            setUser(userData);
            setIsAdmin(userData.role === 'admin');
          } catch (error) {
            console.error('Error parsing user data from storage event:', error);
          }
        }
      }
    };

    window.addEventListener('storage', handleStorageChange);
    return () => window.removeEventListener('storage', handleStorageChange);
  }, [accessToken, router, handleLogout]);

  // Check token expiration
  useEffect(() => {
    if (!accessToken) return;

    const checkTokenExpiration = () => {
      try {
        // JWT tokens are in three parts: header.payload.signature
        const payload = accessToken.split('.')[1];
        if (!payload) return;

        // Decode the base64 payload
        const decodedPayload = JSON.parse(atob(payload));
        const expirationTime = decodedPayload.exp * 1000; // Convert to milliseconds

        if (Date.now() >= expirationTime) {
          console.log('Token expired, logging out');
          handleLogout();
        }
      } catch (error) {
        console.error('Error checking token expiration:', error);
      }
    };

    // Check immediately and then every minute
    checkTokenExpiration();
    const interval = setInterval(checkTokenExpiration, 60000);

    return () => clearInterval(interval);
  }, [accessToken, handleLogout]);

  const login = (token: string, userData: User) => {
    if (!userData.role || userData.role !== 'admin') {
      console.error('Access denied: Admin privileges required');
      return;
    }

    setAccessToken(token);
    setUser(userData);
    setIsAuthenticated(true);
    setIsAdmin(userData.role === 'admin');

    localStorage.setItem('access_token', token);
    localStorage.setItem('user', JSON.stringify(userData));
  };

  const value = {
    isAuthenticated,
    user,
    accessToken,
    login,
    logout: handleLogout,
    isLoading,
    isAdmin,
  };

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export const useAuth = () => {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
};
