import { api } from '@/lib/api';
import { LoginCredentials, RegisterCredentials, AuthResponse, User } from '@/types/auth';

// Remove signIn from next-auth/react

// Define a type for the expected login response from our backend
interface BackendLoginResponse {
  access_token: string;
  refresh_token?: string; // Optional since it's primarily handled by cookies
  user: User; // Assuming a User type is defined or we define one
  // Potentially refresh_token if we decide to store it client-side too, though the cookie handles it
}

// Define a type for the expected register response from our backend
// Assuming registration doesn't automatically log in and return tokens
interface BackendRegisterResponse {
  message: string;
  user: User; // Or just user ID/email
}

interface RefreshResponse {
  access_token: string;
  refresh_token?: string;
}

export class AuthService {
  static async login(credentials: LoginCredentials): Promise<BackendLoginResponse> {
    try {
      const response = await api.post<BackendLoginResponse>('/users/login', {
        email: credentials.email,
        password: credentials.password,
      });

      if (response.data && response.data.access_token) {
        // Store only the access token and user data in localStorage
        localStorage.setItem('accessToken', response.data.access_token);
        localStorage.setItem('user', JSON.stringify(response.data.user));

        // Note: refresh token is handled automatically by the browser via HttpOnly cookie
        return response.data;
      } else {
        throw new Error('Login failed: No access token received.');
      }
    } catch (error: any) {
      console.error('Login error:', error);
      throw new Error(error.response?.data?.message || error.message || 'Login failed');
    }
  }

  static async register(credentials: RegisterCredentials): Promise<BackendRegisterResponse> { // Return backend register response type
    try {
      const registerData = {
        Email: credentials.Email,
        Username: credentials.Username,
        Password: credentials.Password,
        FirstName: credentials.FirstName,
        LastName: credentials.LastName,
        PhoneNumber: credentials.PhoneNumber || ""
      };

      // Make the registration request to our backend
      const response = await api.post<BackendRegisterResponse>('/users/register', registerData);

      // Registration successful, return backend response
      // We won't automatically log in here anymore, user needs to log in separately
      return response.data;

    } catch (error: any) {
      if (error.response?.status === 409) {
        throw new Error(error.response?.data?.message || 'Username or email already exists');
      }
      console.error('Registration error:', error);
      throw new Error(error.response?.data?.message || error.message || 'Registration failed. Please try again.');
    }
  }

  static async refreshToken(): Promise<string> {
    try {
      // The refresh token is automatically included in the request as an HttpOnly cookie
      const response = await api.post<RefreshResponse>('/users/refresh');

      if (response.data && response.data.access_token) {
        localStorage.setItem('accessToken', response.data.access_token);
        return response.data.access_token;
      }
      throw new Error('Token refresh failed');
    } catch (error: any) {
      console.error('Token refresh error:', error);
      // If refresh fails, force logout
      this.logout();
      throw new Error('Session expired. Please login again.');
    }
  }

  static async logout(): Promise<void> {
    try {
      // Clear the refresh token cookie by making a request to the logout endpoint
      await api.post('/users/logout');
    } finally {
      // Always clear localStorage, even if the API call fails
      localStorage.removeItem('accessToken');
      localStorage.removeItem('user');
    }
  }

  static getAccessToken(): string | null {
    return localStorage.getItem('accessToken');
  }

  static getCurrentUser(): User | null {
    const userStr = localStorage.getItem('user');
    return userStr ? JSON.parse(userStr) : null;
  }
}

