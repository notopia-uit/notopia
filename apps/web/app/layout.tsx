import { client } from '@notopia-uit/api-gen/client.gen';
import QueryClientProvider from '@notopia-uit/ui/components/client-query-provider';
import { ThemeProvider } from '@notopia-uit/ui/components/theme-provider';

import './globals.css';
import { fetchAccessTokenServerSide } from '@notopia-uit/ui/lib/get-access-token';
import { cn } from '@notopia-uit/ui/lib/shadcn/utils';
import { Inter as FontSans } from 'next/font/google';
import localFont from 'next/font/local';

client.setConfig({
  auth: fetchAccessTokenServerSide,
});

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
            <QueryClientProvider>{children}</QueryClientProvider>
          </ThemeProvider>
        </body>
      </html>
    </>
  );
}
