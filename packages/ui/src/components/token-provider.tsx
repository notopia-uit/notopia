'use client';

import { client } from '@notopia-uit/api-gen/client.gen';
import { fetchAccessTokenClientSide } from '@notopia-uit/ui/lib/get-access-token-client-side';
import * as React from 'react';

export function ApiProvider({ children }) {
  client.setConfig({
    auth: fetchAccessTokenClientSide,
  });

  return <>{children}</>;
}
