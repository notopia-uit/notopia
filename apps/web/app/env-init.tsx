'use client';

import { configureAuthClient } from '@notopia-uit/ui/lib/auth-client';

let configured = false;

export function EnvInit({
  children,
  betterAuthUrl,
}: {
  children: React.ReactNode;
  betterAuthUrl: string;
}) {
  if (!configured) {
    if (!betterAuthUrl) {
      console.warn(
        'BETTER_AUTH_URL is not set. Falling back to http://localhost:3000'
      );
    }
    configureAuthClient(betterAuthUrl || 'http://localhost:3000');
    configured = true;
  }
  return <>{children}</>;
}
