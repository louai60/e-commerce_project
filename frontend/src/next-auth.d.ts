import NextAuth, { DefaultSession, DefaultUser } from "next-auth";
import { JWT, DefaultJWT } from "next-auth/jwt";

declare module "next-auth" {
  /**
   * Returned by `useSession`, `getSession` and received as a prop on the `SessionProvider` React Context
   */
  interface Session {
    accessToken?: string; // Add accessToken property
    user: {
      id: string; // Add id property
      role: string; // Add role property
    } & DefaultSession["user"]; // Keep default properties like name, email, image
  }

  /**
   * The shape of the user object returned in the OAuth providers' `profile` callback,
   * or the second parameter of the `session` callback, when using a database.
   * Also returned by the `authorize` callback of the Credentials provider.
   */
  interface User extends DefaultUser {
    // Add properties returned by your authorize callback
    accessToken?: string;
    role?: string;
    // refreshToken is NOT included here as it's handled by HttpOnly cookie
  }
}

declare module "next-auth/jwt" {
  /** Returned by the `jwt` callback and `getToken`, when using JWT sessions */
  interface JWT extends DefaultJWT {
    // Add properties added in the jwt callback
    accessToken?: string;
    role?: string;
    id?: string;
    // refreshToken is NOT included here
  }
}
