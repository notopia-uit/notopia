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
    configureAuthClient(betterAuthUrl);
    configured = true;
  }
  return <>{children}</>;
}
