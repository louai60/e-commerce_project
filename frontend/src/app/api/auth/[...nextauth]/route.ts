import NextAuth from 'next-auth';
import CredentialsProvider from 'next-auth/providers/credentials';

const handler = NextAuth({
  providers: [
    CredentialsProvider({
      name: 'Credentials',
      credentials: {
        email: { label: 'Email', type: 'email' },
        password: { label: 'Password', type: 'password' }
      },
      async authorize(credentials) {
        if (!credentials?.email || !credentials?.password) {
          throw new Error('Email and password required');
        }

        try {
          const response = await fetch(`${process.env.NEXT_PUBLIC_API_URL}/users/login`, {
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

          return {
            id: data.user.id,
            email: data.user.email,
            name: `${data.user.firstName} ${data.user.lastName}`,
            accessToken: data.access_token,
            refreshToken: data.refresh_token,
            role: data.user.role,
          };
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
      if (user) {
        token.accessToken = user.accessToken;
        token.refreshToken = user.refreshToken;
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
        session.refreshToken = token.refreshToken as string;
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
