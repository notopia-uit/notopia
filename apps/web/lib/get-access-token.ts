import { headers } from 'next/headers';
import { notFound } from 'next/navigation';

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
    notFound();
  }
  return data.accessToken;
};
