# Admin Dashboard Authentication Implementation

This document outlines the authentication implementation for the admin dashboard, ensuring that only authenticated admin users can access protected routes.

## Overview

The admin dashboard implements a comprehensive authentication system with the following features:

1. **Route Protection**: All admin routes are protected and require authentication
2. **Admin Role Verification**: Only users with the admin role can access the dashboard
3. **Token Expiration Handling**: Expired tokens automatically redirect to the login page
4. **Token Refresh**: Automatic token refresh when tokens expire
5. **Persistent Authentication**: Authentication state is persisted across page refreshes

## Implementation Components

### 1. Authentication Context

The `AuthContext` provides authentication state and functions to the entire application:

- `isAuthenticated`: Boolean indicating if the user is authenticated
- `user`: The current user object
- `accessToken`: The current JWT access token
- `login`: Function to log in a user
- `logout`: Function to log out a user
- `isLoading`: Boolean indicating if authentication state is loading
- `isAdmin`: Boolean indicating if the user has admin role

### 2. Middleware

The Next.js middleware (`middleware.ts`) intercepts all requests and:

- Redirects unauthenticated users to the login page
- Redirects authenticated users away from login/signup pages
- Validates JWT tokens and checks for expiration
- Verifies admin role for protected routes
- Handles token refresh when tokens expire

### 3. API Interceptors

The Axios interceptors in `api.ts` handle:

- Adding authentication tokens to requests
- Refreshing tokens when they expire
- Redirecting to login when refresh fails
- Handling 401 and 403 errors

### 4. Protected Layouts

The admin layout (`(admin)/layout.tsx`) uses the `useAuth` hook to:

- Check authentication status
- Verify admin role
- Show loading state while checking
- Redirect to login if not authenticated

### 5. Authentication Service

The `AuthService` provides methods for:

- `login`: Authenticating users
- `logout`: Logging out users
- `getCurrentUser`: Getting the current user
- `isAuthenticated`: Checking if a user is authenticated
- `isAdmin`: Checking if a user has admin role

## Authentication Flow

1. **Login**:
   - User enters credentials on the login page
   - Credentials are sent to the API
   - API returns JWT token and user data
   - Token and user data are stored in localStorage
   - User is redirected to the dashboard

2. **Route Access**:
   - Middleware checks for token on each route access
   - If token is missing, user is redirected to login
   - If token is present, middleware validates it
   - If token is expired, middleware attempts to refresh it
   - If refresh fails, user is redirected to login

3. **API Requests**:
   - Token is added to all API requests
   - If a request returns 401, token refresh is attempted
   - If refresh succeeds, the original request is retried
   - If refresh fails, user is logged out and redirected to login

4. **Logout**:
   - User clicks logout
   - Token and user data are removed from localStorage
   - Logout request is sent to invalidate refresh token
   - User is redirected to login page

## Security Considerations

1. **Token Storage**:
   - Access token is stored in localStorage for persistence
   - Refresh token is stored as an HttpOnly cookie for security

2. **Token Validation**:
   - Tokens are validated on both client and server
   - Expired tokens are automatically refreshed or rejected

3. **Role-Based Access**:
   - Only users with admin role can access the dashboard
   - Role is verified on both client and server

4. **Token Expiration**:
   - Access tokens have a short lifespan (typically 15 minutes)
   - Refresh tokens have a longer lifespan (typically 7 days)
   - Expired tokens are automatically refreshed when possible

## Testing Authentication

To test the authentication implementation:

1. **Login Test**:
   - Navigate to `/signin`
   - Enter valid admin credentials
   - Verify redirect to dashboard

2. **Protection Test**:
   - Clear localStorage and cookies
   - Attempt to access a protected route
   - Verify redirect to login page

3. **Role Test**:
   - Login with non-admin credentials
   - Attempt to access admin dashboard
   - Verify access is denied

4. **Expiration Test**:
   - Modify a valid token to be expired
   - Attempt to access a protected route
   - Verify token refresh or redirect to login

## Troubleshooting

Common authentication issues and solutions:

1. **"Not authenticated" error**:
   - Check localStorage for access_token
   - Verify token has not expired
   - Check network requests for 401 errors

2. **"Access denied" error**:
   - Verify user has admin role
   - Check for 403 errors in network requests

3. **Redirect loops**:
   - Check middleware configuration
   - Verify token validation logic
   - Clear browser storage and cookies

4. **Token refresh issues**:
   - Check refresh token in cookies
   - Verify refresh endpoint is working
   - Check for network errors during refresh
