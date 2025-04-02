import api from '@/lib/api';
import { LoginCredentials, RegisterCredentials, AuthResponse } from '@/types/auth';
import { signIn } from 'next-auth/react';

export class AuthService {
  static async login(credentials: LoginCredentials): Promise<AuthResponse> {
    try {
      const result = await signIn('credentials', {
        email: credentials.email,
        password: credentials.password,
        redirect: false,
        callbackUrl: '/'
      });

      if (!result) {
        throw new Error('Authentication failed');
      }

      if (result.error) {
        throw new Error(result.error);
      }

      return result as any;
    } catch (error: any) {
      console.error('Login error:', error);
      throw new Error(error.message || 'Service unavailable');
    }
  }

  static async register(credentials: RegisterCredentials): Promise<AuthResponse> {
    try {
      const response = await api.post<AuthResponse>('/users/register', credentials);
      
      const result = await signIn('credentials', {
        email: credentials.Email,
        password: credentials.Password,
        redirect: false,
      });

      if (result?.error) {
        throw new Error(result.error);
      }

      return response.data;
    } catch (error: any) {
      if (error.response?.status === 409) {
        throw new Error('Username or email already exists');
      }
      throw new Error('Registration failed. Please try again.');
    }
  }
}








