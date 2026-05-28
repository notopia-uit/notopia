'use client';

import { configureAuthClient } from '@notopia-uit/ui/lib/auth-client';

const authUrl = process.env.NEXT_PUBLIC_BETTER_AUTH_URL;
if (!authUrl) {
  console.warn(
    'NEXT_PUBLIC_BETTER_AUTH_URL is not set. Falling back to http://localhost:3000'
  );
}
configureAuthClient(authUrl || 'http://localhost:3000');

export function EnvInit({ children }: { children: React.ReactNode }) {
  return <>{children}</>;
}
