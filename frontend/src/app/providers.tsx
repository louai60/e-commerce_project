'use client';

// Remove SessionProvider import
import { Provider as ReduxProvider } from 'react-redux';
import { store } from '@/redux/store';
import { AuthProvider } from '@/contexts/AuthContext';

export function Providers({ children }: { children: React.ReactNode }) {
  return (
    // Remove SessionProvider wrapper
      <ReduxProvider store={store}>
        <AuthProvider>
          {children}
        </AuthProvider>
      </ReduxProvider>
    // Remove SessionProvider wrapper
  );
}
