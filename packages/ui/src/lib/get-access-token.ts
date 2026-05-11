import { auth } from '@notopia-uit/ui/lib/auth';
import { headers } from 'next/headers';

export const fetchAccessTokenClientSide = async (): Promise<string> => {
  const data = await auth.api.getAccessToken({
    body: {
      providerId: 'authentik',
    },
  });
  if (!data?.accessToken) {
    throw new Error('Missing Authentik access token');
  }
  return data.accessToken;
};

export const fetchAccessTokenServerSide = async (): Promise<string> => {
  const data = await auth.api.getAccessToken({
    body: {
      providerId: 'authentik',
    },
    headers: await headers(),
  });
  if (!data?.accessToken) {
    throw new Error('Missing Authentik access token');
  }
  return data.accessToken;
};
