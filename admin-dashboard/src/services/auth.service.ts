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
      // Add detailed request logging
      console.log('Login attempt:', {
        url: `${API_URL}/users/login`,
        email: credentials.email,
        apiUrl: API_URL
      });

      const response = await axios.post<LoginResponse>(
        `${API_URL}/users/login`,
        credentials,
        {
          withCredentials: true,
          headers: {
            'Content-Type': 'application/json',
            'Accept': 'application/json'
          }
        }
      );

      // Add response logging
      console.log('Server response:', {
        status: response.status,
        headers: response.headers,
        data: response.data
      });

      if (!response.data) {
        throw new Error('No data received from server');
      }

      if (!response.data.access_token) {
        throw new Error('No access token received');
      }

      // Store access token in both localStorage and cookie
      localStorage.setItem('access_token', response.data.access_token);
      if (response.data.user) {
        localStorage.setItem('user', JSON.stringify(response.data.user));
      }

      // Set cookie with secure flags
      const secure = window.location.protocol === 'https:' ? 'Secure;' : '';
      document.cookie = `access_token=${response.data.access_token}; path=/; ${secure} SameSite=Lax; max-age=86400`;

      return response.data;
    } catch (error: any) {
      // Enhanced error logging
      console.error('Login error details:', {
        message: error.message,
        response: error.response?.data,
        status: error.response?.status,
        config: error.config
      });

      if (error.response?.status === 401) {
        throw new Error('Invalid email or password');
      } else if (error.response?.status === 403) {
        throw new Error('Access denied: Admin privileges required');
      } else if (error.response?.data?.message) {
        throw new Error(error.response.data.message);
      } else if (error.message) {
        throw new Error(`Login failed: ${error.message}`);
      }
      
      throw new Error('Unable to connect to the server. Please try again later.');
    }
  }

  static logout(): void {
    localStorage.removeItem('access_token');
    localStorage.removeItem('user');

    // Clear the cookie
    document.cookie = 'access_token=; path=/; expires=Thu, 01 Jan 1970 00:00:00 GMT';

    // Call logout endpoint to invalidate refresh token
    axios.post(`${API_URL}/users/logout`, {}, { withCredentials: true })
      .catch(error => console.error('Logout error:', error));
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

