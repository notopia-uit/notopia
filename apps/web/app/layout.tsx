import QueryClientProvider from '@notopia-uit/ui/components/client-query-provider';
import { ThemeProvider } from '@notopia-uit/ui/components/theme-provider';
import { ApiProvider } from '@notopia-uit/ui/components/token-provider';
import { MeilisearchProvider } from '@notopia-uit/ui/contexts/meilisearch-context';

import './globals.css';
import { cn } from '@notopia-uit/ui/lib/shadcn/utils';
import { Inter as FontSans } from 'next/font/google';
import localFont from 'next/font/local';

const fontSans = FontSans({
  subsets: ['latin'],
  variable: '--font-sans',
});

const fontHeading = localFont({
  src: '../assets/fonts/CalSans-SemiBold.woff2',
  variable: '--font-heading',
});

interface RootLayoutProps {
  children: React.ReactNode;
}

export default function RootLayout({ children }: RootLayoutProps) {
  const meilisearchHost = process.env.NEXT_PUBLIC_MEILISEARCH_HOST || 'http://localhost:7700';

  return (
    <>
      <html lang="en" suppressHydrationWarning>
        <head />
        <body
          className={cn(
            'min-h-screen bg-background font-sans antialiased',
            fontSans.variable,
            fontHeading.variable
          )}
        >
          <ThemeProvider
            attribute="class"
            defaultTheme="system"
            enableSystem
            disableTransitionOnChange
            scriptProps={{ type: 'application/json' }}
          >
            <QueryClientProvider>
              <ApiProvider>
                <MeilisearchProvider host={meilisearchHost}>
                  {children}
                </MeilisearchProvider>
              </ApiProvider>
            </QueryClientProvider>
          </ThemeProvider>
        </body>
      </html>
    </>
  );
}
