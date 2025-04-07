import NextAuth, { User } from 'next-auth'; // Import User type
import CredentialsProvider from 'next-auth/providers/credentials';

const handler = NextAuth({
  providers: [
    CredentialsProvider({
      name: 'Credentials',
      credentials: {
        email: { label: 'Email', type: 'email' },
        password: { label: 'Password', type: 'password' }
      },
      async authorize(credentials): Promise<User | null> {
        if (!credentials?.email || !credentials?.password) {
          throw new Error('Email and password required');
        }

        try {
          const response = await fetch(`${process.env.NEXT_PUBLIC_API_URL}/api/v1/users/login`, {
            method: 'POST',
            headers: {
              'Content-Type': 'application/json',
            },
            body: JSON.stringify({
              email: credentials.email,
              password: credentials.password,
            }),
          });

          const data = await response.json();

          if (!response.ok) {
            throw new Error(data.error || 'Authentication failed');
          }

          // The function's return type annotation should handle this now
          const userPayload = {
            id: String(data.user.userId), // Ensure id is a string
            email: data.user.email,
            name: `${data.user.firstName} ${data.user.lastName}`,
            accessToken: data.access_token,
            role: data.user.role,
            // No refreshToken property here
          };

          return userPayload as User; // Assert the type here

        } catch (error: any) {
          throw new Error(error.message || 'Authentication failed');
        }
      }
    })
  ],
  pages: {
    signIn: '/signin',
    error: '/signin'
  },
  callbacks: {
    async jwt({ token, user }) {
      // The 'user' object comes from the 'authorize' callback
      if (user) {
        token.accessToken = user.accessToken;
        // token.refreshToken = user.refreshToken; // Removed
        token.role = user.role;
        token.id = user.id;
      }
      return token;
    },
    async session({ session, token }) {
      if (token) {
        session.user = {
          ...session.user,
          id: token.id as string,
          role: token.role as string
        };
        session.accessToken = token.accessToken as string;
        // session.refreshToken = token.refreshToken as string; // Removed
      }
      return session;
    }
  },
  session: {
    strategy: 'jwt',
  },
  debug: process.env.NODE_ENV === 'development',
});

export { handler as GET, handler as POST };
