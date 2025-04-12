import axios, { AxiosError, InternalAxiosRequestConfig } from 'axios';
// Remove getSession, signOut
import { AuthService } from '@/services/auth.service'; // Import for logout

let isRefreshing = false;
let failedQueue: any[] = [];

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
  baseURL: process.env.NEXT_PUBLIC_API_URL,
  headers: {
    'Content-Type': 'application/json',
  },
  withCredentials: true, // Important for cookies
});

// Request interceptor
api.interceptors.request.use(
  (config) => {
    // Read token from localStorage instead of session
    const token = typeof window !== 'undefined' ? localStorage.getItem('accessToken') : null;
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => Promise.reject(error)
);

// Response interceptor
api.interceptors.response.use(
  (response) => response,
  async (error: AxiosError) => {
    const originalRequest = error.config as InternalAxiosRequestConfig & { _retry?: boolean };
    
    // Check for 401 Unauthorized and ensure it's not a retry
    if (error.response?.status === 401 && !originalRequest._retry) {
      if (isRefreshing) {
        // If already refreshing, queue the request
        return new Promise((resolve, reject) => {
          failedQueue.push({ resolve, reject });
        })
          .then(token => {
            if (originalRequest.headers) { // Check if headers exist
                 originalRequest.headers.Authorization = `Bearer ${token}`;
            }
            return api(originalRequest);
          })
          .catch(err => Promise.reject(err));
      }

      originalRequest._retry = true; // Mark as retry
      isRefreshing = true;

      try {
        console.log('Attempting token refresh...');
        // Attempt to refresh the token using the refresh endpoint
        const response = await api.post('/users/refresh', {}, {
          withCredentials: true // Ensure cookies (refresh_token) are sent
        });

        const { access_token: newAccessToken } = response.data;
        console.log('Token refresh successful.');

        if (!newAccessToken) {
          throw new Error('No new access token received during refresh');
        }

        // Store the new token in localStorage
        localStorage.setItem('accessToken', newAccessToken);

        // Update the Authorization header for the original request
         if (originalRequest.headers) { // Check if headers exist
             originalRequest.headers.Authorization = `Bearer ${newAccessToken}`;
         }

        // Process the queue with the new token
        processQueue(null, newAccessToken);

        // Retry the original request with the new token
        return api(originalRequest);

      } catch (refreshError: any) {
        console.error('Token refresh failed:', refreshError);
        processQueue(refreshError, null); // Reject queued requests

        // Logout the user: clear storage and redirect
        AuthService.logout(); // Use the logout method to clear storage
        // Redirect to sign-in page (client-side only)
        if (typeof window !== 'undefined') {
            window.location.href = '/signin'; // Force redirect
        }

        return Promise.reject(refreshError);
      } finally {
        isRefreshing = false;
      }
    }
    
    return Promise.reject(error);
  }
);


