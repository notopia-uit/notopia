import { authClient } from '@notopia-uit/ui/lib/auth-client';

export const fetchAccessToken = async (): Promise<string> => {
  const { data, error } = await authClient.getAccessToken({
    providerId: 'authentik',
  });
  if (error || !data?.accessToken) {
    throw new Error('Missing Authentik access token');
  }
  return data.accessToken;
};
