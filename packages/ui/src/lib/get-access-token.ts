import { auth } from '@notopia-uit/ui/lib/auth';
import { headers } from 'next/headers';

export const fetchAccessTokenServerSide = async (): Promise<string> => {
  const h = await headers();
  const data = await auth.api.getAccessToken({
    body: {
      providerId: 'authentik',
    },
    headers: h,
  });
  if (!data?.accessToken) {
    throw new Error('Missing Authentik access token');
  }
  return data.accessToken;
};
