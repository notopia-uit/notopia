'use client';
import { getAuthClient } from '@notopia-uit/ui/lib/auth-client';

export const fetchAccessTokenClientSide = async (): Promise<string> => {
  const client = getAuthClient();
  const { data: accounts } = await client.listAccounts();
  const account = accounts?.find((a) => a.providerId === 'authentik');
  if (!account) {
    throw new Error('Missing Authentik account for current user');
  }
  const data = await client.getAccessToken({
    accountId: account.id,
  });
  if (!data?.data?.accessToken) {
    throw new Error('Missing Authentik access token from client side fetch');
  }
  return data.data.accessToken;
};
