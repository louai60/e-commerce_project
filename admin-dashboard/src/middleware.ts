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
    // redirect to signin
    console.log('Redirecting unauthenticated user to signin');
    return NextResponse.redirect(new URL('/signin', request.url));
  }

  // For all other cases, continue with the request
  return NextResponse.next();
}

// Configure the paths that should be handled by this middleware
export const config = {
  matcher: [
    '/dashboard/:path*',
    '/signin',
    '/signup',
  ],
};