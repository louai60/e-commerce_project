// Type definitions for Next.js API routes
import { NextRequest } from 'next/server';

declare module 'next/server' {
  // Define the params type for API routes
  export interface RouteParams {
    params: {
      [key: string]: string;
    };
  }

  // Extend the GET handler type
  export type RouteHandler = (
    request: NextRequest,
    context: RouteParams
  ) => Promise<Response> | Response;
}
