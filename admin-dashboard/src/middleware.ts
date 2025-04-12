import { NextResponse } from 'next/server';
import type { NextRequest } from 'next/server';

export function middleware(request: NextRequest) {
  // Get the pathname
  const path = request.nextUrl.pathname;

  // Define public paths that don't require authentication
  const isPublicPath = path === '/signin' || path === '/signup';

  // Check if user is authenticated by looking for the access token
  const token = request.cookies.get('access_token')?.value;

  if (isPublicPath && token) {
    // If user is authenticated and tries to access public path,
    // redirect to dashboard
    return NextResponse.redirect(new URL('/dashboard', request.url));
  }

  if (!isPublicPath && !token) {
    // If user is not authenticated and tries to access protected path,
    // redirect to signin
    return NextResponse.redirect(new URL('/signin', request.url));
  }
}

// Configure the paths that should be handled by this middleware
export const config = {
  matcher: [
    '/dashboard/:path*',
    '/signin',
    '/signup',
  ],
};