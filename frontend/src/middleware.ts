import { NextResponse } from 'next/server';
import type { NextRequest } from 'next/server';

export async function middleware(request: NextRequest) {
  // Middleware runs server-side (Edge) and cannot directly access localStorage.
  // The current setup relies on client-side localStorage via `api.ts`.
  // Authentication checks for protected routes should primarily happen client-side
  // based on the presence/validity of the token in localStorage, or by handling
  // API errors (like 401s) which the `api.ts` interceptor already does.

  // We will remove the checks here and rely on client-side logic or API error handling.
  // If server-side protection is strictly needed, a session mechanism accessible
  // by middleware (like NextAuth.js with JWT cookies) would be required.

  // TODO: Implement proper server-side session checking if required, potentially using NextAuth.js.

  // Allow all requests to pass through middleware for now.
  // Client-side logic or API interceptors will handle auth redirects.
  return NextResponse.next();
}

// Keep the matcher config if you intend to add checks back later,
// otherwise, you could remove it if middleware isn't doing anything.
// For now, we leave it to indicate these paths *should* be protected.
export const config = {
  matcher: [
    '/profile/:path*',
    '/dashboard/:path*',
    '/orders/:path*',
    '/settings/:path*',
    // Add other paths that require authentication
  ],
};
