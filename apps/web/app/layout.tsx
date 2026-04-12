import QueryClientProvider from '@notopia-uit/ui/components/client-query-provider';
import { ThemeProvider } from '@notopia-uit/ui/components/theme-provider';
import { cn } from '@notopia-uit/ui/lib/utils';
import { Inter as FontSans } from 'next/font/google';
import localFont from 'next/font/local';

import './globals.css';

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
          >
            <QueryClientProvider>{children}</QueryClientProvider>
          </ThemeProvider>
        </body>
      </html>
    </>
  );
}
