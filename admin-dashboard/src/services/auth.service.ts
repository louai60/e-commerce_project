import axios from 'axios';

const API_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api/v1';

interface LoginCredentials {
  email: string;
  password: string;
}

interface User {
  id: number;
  email: string;
  username: string;
  role: string;
  // Add other user fields as needed
}

interface LoginResponse {
  access_token: string;
  user: User;
}

interface RegisterData {
  firstName: string;
  lastName: string;
  email: string;
  password: string;
}

export class AuthService {
  static async login(credentials: LoginCredentials): Promise<LoginResponse> {
    try {
      const response = await axios.post<LoginResponse>(
        `${API_URL}/users/login`,
        credentials,
        { withCredentials: true } // Important for handling refresh token cookie
      );

      if (response.data && response.data.access_token) {
        // Store access token and user data
        localStorage.setItem('access_token', response.data.access_token);
        localStorage.setItem('user', JSON.stringify(response.data.user));
        return response.data;
      }
      throw new Error('Login failed: Invalid response format');
    } catch (error: any) {
      if (error.response?.status === 401) {
        throw new Error('Invalid credentials');
      }
      throw new Error(error.response?.data?.error || 'Login failed');
    }
  }

  static logout(): void {
    localStorage.removeItem('access_token');
    localStorage.removeItem('user');
    // Optionally call logout endpoint to invalidate refresh token
    axios.post(`${API_URL}/users/logout`, {}, { withCredentials: true })
      .catch(console.error);
  }

  static getCurrentUser(): User | null {
    const userStr = localStorage.getItem('user');
    if (!userStr) return null;
    try {
      const user = JSON.parse(userStr);
      return user;
    } catch {
      return null;
    }
  }

  static isAuthenticated(): boolean {
    return !!localStorage.getItem('access_token');
  }

  static isAdmin(): boolean {
    const user = this.getCurrentUser();
    return user?.role === 'admin';
  }

  static async register(data: RegisterData, headers?: Record<string, string>): Promise<void> {
    try {
      const response = await fetch(`${process.env.NEXT_PUBLIC_API_URL}/users/admin`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          ...headers
        },
        body: JSON.stringify(data)
      });

      if (!response.ok) {
        const error = await response.json();
        throw new Error(error.error || 'Registration failed');
      }

      return response.json();
    } catch (error: any) {
      if (error.response?.data?.message) {
        throw new Error(error.response.data.message);
      }
      if (error.response?.status === 409) {
        throw new Error('Email or username already exists');
      }
      throw new Error('Registration failed. Please try again later.');
    }
  }
}

