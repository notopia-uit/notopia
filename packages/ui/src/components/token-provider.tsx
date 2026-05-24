'use client';

import { client } from '@notopia-uit/api-gen';
import { fetchAccessTokenClientSide } from '@notopia-uit/ui/lib/get-access-token-client-side';
import * as React from 'react';
import { AlertProvider } from './alert-context';
import { AlertContainer } from './alert-container';

export function ApiProvider({ children }) {
  client.setConfig({
    auth: fetchAccessTokenClientSide,
  });

  return (
    <AlertProvider>
      {children}
      <AlertContainer />
    </AlertProvider>
  );
}
