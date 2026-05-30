import { headers } from 'next/headers';

import { auth } from './auth';

export const fetchAccessTokenServerSide = async (): Promise<string> => {
  const h = await headers();
  const data = await auth.api.getAccessToken({
    body: {
      providerId: 'authentik',
    },
    headers: h,
  });
  if (!data?.accessToken) {
    throw new Error('Missing Authentik access token from server side fetch');
  }
  return data.accessToken;
};
