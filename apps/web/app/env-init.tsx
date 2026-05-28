'use client';

import { configureAuthClient } from '@notopia-uit/ui/lib/auth-client';

configureAuthClient(
  process.env.NEXT_PUBLIC_BETTER_AUTH_URL || 'http://localhost:3000'
);

export function EnvInit({ children }: { children: React.ReactNode }) {
  return <>{children}</>;
}
