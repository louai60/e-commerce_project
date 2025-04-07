import axios, { AxiosError, InternalAxiosRequestConfig } from 'axios';
import { getSession, signIn, signOut } from 'next-auth/react';

// Flag to prevent infinite refresh loops
let isRefreshing = false;
let failedQueue: { resolve: (value: unknown) => void; reject: (reason?: any) => void; }[] = [];

const processQueue = (error: AxiosError | null, token: string | null = null) => {
  failedQueue.forEach(prom => {
    if (error) {
      prom.reject(error);
    } else {
      prom.resolve(token);
    }
  });
  failedQueue = [];
};


const api = axios.create({
  // Remove /api/v1 from the baseURL as it's already part of the routes
  baseURL: process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080',
  headers: {
    'Content-Type': 'application/json',
  },
  withCredentials: true,
});

// Request interceptor for adding auth token
api.interceptors.request.use(async (config) => {
  const session = await getSession();
  if (session?.accessToken) {
    config.headers.Authorization = `Bearer ${session.accessToken}`;
  }
  return config;
}, (error) => {
  return Promise.reject(error);
});

// Response interceptor for handling token refresh
api.interceptors.response.use(
  (response) => {
    // Any status code that lie within the range of 2xx cause this function to trigger
    return response;
  },
  async (error: AxiosError) => {
    const originalRequest = error.config as InternalAxiosRequestConfig & { _retry?: boolean };

    // Check if it's a 401 error and not a retry request already
    if (error.response?.status === 401 && !originalRequest._retry) {

      // Prevent multiple refresh calls simultaneously
      if (isRefreshing) {
        // Add the original request to a queue to be retried later
        return new Promise(function(resolve, reject) {
          failedQueue.push({ resolve, reject });
        }).then(token => {
          if (originalRequest.headers) {
            originalRequest.headers['Authorization'] = 'Bearer ' + token;
          }
          return api(originalRequest);
        }).catch(err => {
          return Promise.reject(err); // Propagate the error if refresh fails
        });
      }

      originalRequest._retry = true; // Mark as retry
      isRefreshing = true;

      try {
        console.log('Attempting token refresh...');
        // Use a separate axios instance or fetch for refresh to avoid interceptor loop
        const refreshResponse = await axios.post(
          `${process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'}/users/refresh`,
          {}, // No body needed, refresh token is in cookie
          { withCredentials: true } // Ensure cookies are sent
        );

        const { access_token: newAccessToken } = refreshResponse.data;

        if (!newAccessToken) {
          throw new Error('No new access token received');
        }

        console.log('Token refresh successful.');

        // Update the session - This is tricky with NextAuth client-side.
        // Re-triggering signIn might be the most reliable way, though not ideal.
        // Alternatively, force a session refetch, hoping it picks up the new cookie implicitly?
        // Update the original request header
        if (originalRequest.headers) {
          originalRequest.headers['Authorization'] = `Bearer ${newAccessToken}`;
        }

        // Process the queue with the new token
        processQueue(null, newAccessToken);

        // Retry the original request
        return api(originalRequest);

      } catch (refreshError: any) {
        console.error('Token refresh failed:', refreshError);
        // Process the queue with the error
        processQueue(refreshError, null);
        // Sign out the user if refresh fails
        await signOut({ redirect: false }); // Avoid redirect loop if signout page needs auth
        return Promise.reject(refreshError);
      } finally {
        isRefreshing = false;
      }
    }

    // For errors other than 401 or if it's already a retry, reject the promise
    return Promise.reject(error);
  }
);


export default api;

