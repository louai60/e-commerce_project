import { Outfit } from 'next/font/google';
import './globals.css';

import { SidebarProvider } from '@/context/SidebarContext';
import { ThemeProvider } from '@/context/ThemeContext';
import { SWRConfig } from 'swr';
import { swrConfig } from '@/lib/swr-config';
import { ApolloProvider } from '@/components/providers/ApolloProvider';
import { AuthProvider } from '@/contexts/AuthContext';

const outfit = Outfit({
  subsets: ["latin"],
});

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en">
      <body className={`${outfit.className} dark:bg-gray-900`}>
        <SWRConfig value={swrConfig}>
          <ApolloProvider>
            <AuthProvider>
              <ThemeProvider>
                <SidebarProvider>{children}</SidebarProvider>
              </ThemeProvider>
            </AuthProvider>
          </ApolloProvider>
        </SWRConfig>
      </body>
    </html>
  );
}
