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
      const registerData = {
        Email: credentials.Email,
        Username: credentials.Username,
        Password: credentials.Password,
        FirstName: credentials.FirstName,
        LastName: credentials.LastName,
        PhoneNumber: credentials.PhoneNumber || "" // Optional field
      };

      // Make the registration request
      const response = await api.post<AuthResponse>('/api/v1/users/register', registerData);
      
      // Attempt to sign in immediately after successful registration
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
      console.error('Registration error:', error);
      throw new Error(error.message || 'Registration failed. Please try again.');
    }
  }
}

