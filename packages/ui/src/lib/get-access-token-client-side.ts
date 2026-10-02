'use client';
import { getAuthClient } from '@notopia-uit/ui/lib/auth-client';

export const fetchAccessTokenClientSide = async (): Promise<string> => {
  const client = getAuthClient();
  const data = await client.getAccessToken({
    useAccountCookie: true,
  });
  if (!data?.data?.accessToken) {
    throw new Error('Missing Authentik access token from client side fetch');
  }
  return data.data.accessToken;
};
