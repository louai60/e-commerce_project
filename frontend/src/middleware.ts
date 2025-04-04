import { NextResponse } from 'next/server';
import type { NextRequest } from 'next/server';
import { getToken } from 'next-auth/jwt';

export async function middleware(request: NextRequest) {
  const { pathname } = request.nextUrl;
  
  // Skip middleware for OAuth callback routes
  if (pathname.startsWith('/api/auth/callback')) {
    return NextResponse.next();
  }

  // Skip middleware for authentication routes
  if (pathname.startsWith('/signin') || pathname.startsWith('/signup')) {
    return NextResponse.next();
  }
  
  // Check if the path is a protected route
  const isProtectedRoute = [
    '/dashboard',
    '/profile',
    '/orders',
    '/cart',
    '/checkout',
  ].some(route => pathname.startsWith(route));
  
  // Get the token from the session
  const token = await getToken({
    req: request,
    secret: process.env.NEXTAUTH_SECRET,
  });
  
  // Redirect logic for protected routes only
  if (isProtectedRoute && !token) {
    const url = new URL('/signin', request.url);
    url.searchParams.set('callbackUrl', encodeURI(pathname));
    return NextResponse.redirect(url);
  }
  
  return NextResponse.next();
}

export const config = {
  matcher: [
    // Exclude OAuth callback paths and static assets
    '/((?!api/auth/callback|_next/static|_next/image|favicon.ico|images|.*\\.png$).*)',
  ],
};
