'use client';
import { authClient } from '@notopia-uit/ui/lib/auth-client';

export const fetchAccessTokenClientSide = async (): Promise<string> => {
  const data = await authClient.getAccessToken({
    providerId: 'authentik',
  });
  if (!data?.data?.accessToken) {
    throw new Error('Missing Authentik access token from client side fetch');
  }
  return data.data.accessToken;
};
