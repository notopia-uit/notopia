import { headers } from 'next/headers';
import { notFound } from 'next/navigation';

import { auth } from './auth';

export const fetchAccessTokenServerSide = async (): Promise<string> => {
  const h = await headers();

  const accounts = await auth.api.listUserAccounts({ headers: h });
  const account = accounts?.find((a) => a.providerId === 'authentik');
  if (!account) {
    notFound();
  }
  const data = await auth.api.getAccessToken({
    body: {
      accountId: account.id,
    },
    headers: h,
  });
  if (!data?.accessToken) {
    notFound();
  }
  return data.accessToken;
};
