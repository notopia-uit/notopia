import { headers } from 'next/headers';
import { notFound } from 'next/navigation';

import { auth } from './auth';

export const fetchAccessTokenServerSide = async (): Promise<string> => {
  const h = await headers();

  // Stateless flow (no database): resolve the OAuth account from the signed
  // `account_data` cookie instead of a DB row ID. getAccessToken refreshes
  // the access token via the provider when expired (preserving the refresh
  // token per better-auth#8001) and re-persists the cookie.
  const data = await auth.api.getAccessToken({
    body: {
      useAccountCookie: true,
    },
    headers: h,
  });
  if (!data?.accessToken) {
    notFound();
  }
  return data.accessToken;
};
