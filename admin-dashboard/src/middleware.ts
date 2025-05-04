import { NextResponse } from 'next/server';
import type { NextRequest } from 'next/server';

export function middleware(request: NextRequest) {
  // Get the pathname
  const path = request.nextUrl.pathname;

  // Define public paths that don't require authentication
  const isPublicPath = path === '/signin' || path === '/signup';

  // Check if user is authenticated by looking for the access token
  // Try cookie first, then check Authorization header as fallback
  const token = request.cookies.get('access_token')?.value ||
                request.headers.get('Authorization')?.replace('Bearer ', '');

  console.log(`Middleware: Path=${path}, isPublicPath=${isPublicPath}, hasToken=${!!token}`);

  if (isPublicPath && token) {
    // If user is authenticated and tries to access public path,
    // redirect to home page
    console.log('Redirecting authenticated user to home page');
    return NextResponse.redirect(new URL('/', request.url));
  }

  if (!isPublicPath && !token) {
    // If user is not authenticated and tries to access protected path,
    // redirect to signin with callback URL
    console.log('Redirecting unauthenticated user to signin');
    const callbackUrl = encodeURIComponent(request.nextUrl.pathname);
    return NextResponse.redirect(new URL(`/signin?callbackUrl=${callbackUrl}`, request.url));
  }

  // Check token validity if present
  if (token) {
    try {
      // JWT tokens are in three parts: header.payload.signature
      const parts = token.split('.');
      if (parts.length !== 3) {
        throw new Error('Invalid token format');
      }

      // Decode the base64 payload
      const payload = JSON.parse(
        Buffer.from(parts[1], 'base64').toString()
      );

      // Check if token is expired
      const expirationTime = payload.exp * 1000; // Convert to milliseconds
      if (Date.now() >= expirationTime) {
        console.log('Token expired, redirecting to signin');
        // Clear the invalid token cookie
        const response = NextResponse.redirect(new URL('/signin', request.url));
        response.cookies.delete('access_token');
        return response;
      }

      // Check if user has admin role
      if (!isPublicPath && payload.role !== 'admin') {
        console.log('Access denied: Admin privileges required');
        return NextResponse.redirect(new URL('/signin', request.url));
      }
    } catch (error) {
      console.error('Error validating token:', error);
      // If token is invalid, redirect to signin
      const response = NextResponse.redirect(new URL('/signin', request.url));
      response.cookies.delete('access_token');
      return response;
    }
  }

  // For all other cases, continue with the request
  return NextResponse.next();
}

// Configure the paths that should be handled by this middleware
export const config = {
  matcher: [
    '/((?!api|_next/static|_next/image|favicon.ico).*)',
  ],
};