import { auth } from '@notopia-uit/ui/lib/auth';

export const fetchAccessTokenServerSide = async (): Promise<string> => {
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
