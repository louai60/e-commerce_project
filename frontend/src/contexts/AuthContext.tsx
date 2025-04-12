'use client';

import { createContext, useContext, useState, useEffect, ReactNode, useCallback } from 'react';
import { User } from '@/types/auth'; // Assuming User type exists
import { AuthService } from '@/services/auth.service'; // Import AuthService for logout helper

interface AuthContextType {
  isAuthenticated: boolean;
  accessToken: string | null;
  user: User | null;
  login: (token: string, userData: User) => void;
  logout: () => void;
  isLoading: boolean; // Add loading state for initial check
}

// Default context value matching the type
const AuthContext = createContext<AuthContextType>({
  isAuthenticated: false,
  accessToken: null,
  user: null,
  login: () => { throw new Error('Login function not implemented'); },
  logout: () => { throw new Error('Logout function not implemented'); },
  isLoading: true, // Start in loading state
});

export function AuthProvider({ children }: { children: ReactNode }) {
  const [accessToken, setAccessToken] = useState<string | null>(null);
  const [user, setUser] = useState<User | null>(null);
  const [isAuthenticated, setIsAuthenticated] = useState<boolean>(false);
  const [isLoading, setIsLoading] = useState<boolean>(true);

  // Check localStorage on initial mount (client-side only)
  useEffect(() => {
    const initializeAuth = () => {
      try {
        const storedToken = localStorage.getItem('accessToken');
        const storedUser = localStorage.getItem('user');
        
        if (storedToken && storedUser) {
          setAccessToken(storedToken);
          setUser(JSON.parse(storedUser));
          setIsAuthenticated(true);
        }
      } catch (error) {
        console.error("Error reading auth state from localStorage", error);
        localStorage.removeItem('accessToken');
        localStorage.removeItem('user');
      }
      setIsLoading(false); // Always set loading to false after checking
    };

    // Only run on client side
    if (typeof window !== 'undefined') {
      initializeAuth();
    } else {
      setIsLoading(false); // Don't keep loading on server side
    }
  }, []);

  const login = useCallback((token: string, userData: User) => {
    try {
      localStorage.setItem('accessToken', token);
      localStorage.setItem('user', JSON.stringify(userData));
      setAccessToken(token);
      setUser(userData);
      setIsAuthenticated(true);
      // Add this event listener for any components that need to react to auth changes
      window.dispatchEvent(new Event('storage'));
    } catch (error) {
      console.error("Error saving auth state to localStorage", error);
    }
  }, []);

  const logout = useCallback(async () => {
    try {
      await AuthService.logout(); // This will clear the refresh token cookie
      setAccessToken(null);
      setUser(null);
      setIsAuthenticated(false);
      // Add this event listener for any components that need to react to auth changes
      window.dispatchEvent(new Event('storage'));
    } catch (error) {
      console.error('Error during logout:', error);
      // Still clear local state even if the API call fails
      setAccessToken(null);
      setUser(null);
      setIsAuthenticated(false);
    }
  }, []);

  // Add this effect to listen for storage events from other tabs
  useEffect(() => {
    const handleStorageChange = () => {
      const storedToken = localStorage.getItem('accessToken');
      const storedUser = localStorage.getItem('user');
      
      if (storedToken && storedUser) {
        setAccessToken(storedToken);
        setUser(JSON.parse(storedUser));
        setIsAuthenticated(true);
      } else {
        setAccessToken(null);
        setUser(null);
        setIsAuthenticated(false);
      }
    };

    window.addEventListener('storage', handleStorageChange);
    return () => window.removeEventListener('storage', handleStorageChange);
  }, []);

  const value = {
    isAuthenticated,
    accessToken,
    user,
    login,
    logout,
    isLoading,
  };

  // Remove the loading check that was preventing render
  return (
    <AuthContext.Provider value={value}>
      {children}
    </AuthContext.Provider>
  );
}

export const useAuth = () => useContext(AuthContext);
