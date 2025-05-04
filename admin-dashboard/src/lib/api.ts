import axios, { AxiosError, InternalAxiosRequestConfig } from 'axios';
import { AuthService } from '@/services/auth.service';

const API_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api/v1';

let isRefreshing = false;
let failedQueue: Array<{
  resolve: (value: unknown) => void;
  reject: (reason?: any) => void;
}> = [];

const processQueue = (error: any, token: string | null = null) => {
  failedQueue.forEach(prom => {
    if (error) {
      prom.reject(error);
    } else {
      prom.resolve(token);
    }
  });
  failedQueue = [];
};

export const api = axios.create({
  baseURL: API_URL,
  headers: {
    'Content-Type': 'application/json',
    'Accept': 'application/json'
  },
  withCredentials: true,
});

// Request interceptor
api.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('access_token');
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }

    // Log request details in development only
    if (process.env.NODE_ENV === 'development') {
      console.log('API Request:', {
        method: config.method?.toUpperCase(),
        url: config.url,
        data: config.data,
        headers: config.headers
      });
    }

    return config;
  },
  (error) => {
    console.error('Request error:', error);
    return Promise.reject(error);
  }
);

// Response interceptor
api.interceptors.response.use(
  (response) => {
    // Log successful response in development only
    if (process.env.NODE_ENV === 'development') {
      console.log('API Response:', {
        status: response.status,
        url: response.config.url,
        method: response.config.method?.toUpperCase(),
        data: response.data
      });
    }

    // Ensure the response has the expected structure for list endpoints
    if (response.config.url?.includes('/products') && !response.config.url?.includes('/products/')) {
      // This is a product listing endpoint
      if (!response.data.products) {
        console.warn('Products endpoint did not return expected structure', response.data);
        // Try to fix the response structure
        if (Array.isArray(response.data)) {
          response.data = {
            products: response.data,
            total: response.data.length,
            pagination: {
              current_page: 1,
              total_pages: 1,
              per_page: response.data.length,
              total_items: response.data.length
            }
          };
        }
      }
    }

    return response;
  },
  async (error: AxiosError) => {
    const originalRequest = error.config as InternalAxiosRequestConfig & { _retry?: boolean };

    // Log error response in development only
    if (process.env.NODE_ENV === 'development') {
      console.error('API Error:', {
        status: error.response?.status,
        data: error.response?.data,
        config: error.config
      });
    }

    // Handle 401 Unauthorized errors (token expired or invalid)
    if (error.response?.status === 401 && !originalRequest._retry) {
      if (isRefreshing) {
        // If a token refresh is already in progress, queue this request
        return new Promise((resolve, reject) => {
          failedQueue.push({ resolve, reject });
        })
          .then((token) => {
            originalRequest.headers.Authorization = `Bearer ${token}`;
            return api(originalRequest);
          })
          .catch((err) => {
            // If refresh ultimately fails, redirect to login
            AuthService.logout();
            window.location.href = '/signin';
            return Promise.reject(err);
          });
      }

      originalRequest._retry = true;
      isRefreshing = true;

      try {
        // Attempt to refresh the token
        const response = await api.post('/users/refresh');
        const { access_token } = response.data;

        if (!access_token) {
          throw new Error('No access token received during refresh');
        }

        // Update token in localStorage
        localStorage.setItem('access_token', access_token);

        // Update the user object if it's in the response
        if (response.data.user) {
          localStorage.setItem('user', JSON.stringify(response.data.user));
        }

        // Update Authorization header for the original request
        originalRequest.headers.Authorization = `Bearer ${access_token}`;

        // Process all queued requests with the new token
        processQueue(null, access_token);

        // Retry the original request with the new token
        return api(originalRequest);
      } catch (refreshError) {
        // Token refresh failed
        processQueue(refreshError, null);

        // Clear auth data and redirect to login
        AuthService.logout();

        // Only redirect in browser environment
        if (typeof window !== 'undefined') {
          const currentPath = window.location.pathname;
          window.location.href = `/signin?callbackUrl=${encodeURIComponent(currentPath)}`;
        }

        return Promise.reject(refreshError);
      } finally {
        isRefreshing = false;
      }
    }

    // Handle 403 Forbidden errors (insufficient permissions)
    if (error.response?.status === 403) {
      console.error('Access denied: Insufficient permissions');

      // Only redirect in browser environment
      if (typeof window !== 'undefined' && !window.location.pathname.includes('/signin')) {
        window.location.href = '/signin';
      }
    }

    return Promise.reject(error);
  }
);