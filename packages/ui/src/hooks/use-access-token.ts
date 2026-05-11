'use client';

import { fetchAccessTokenClientSide } from '@notopia-uit/ui/lib/get-access-token';
import { useEffect, useState } from 'react';

export function useAccessToken() {
  const [accessToken, setAccessToken] = useState<string | null>(null);

  useEffect(() => {
    const getAccessToken = async () => {
      const token = await fetchAccessTokenClientSide();
      setAccessToken(token);
    };
    getAccessToken();
  }, []);

  return accessToken;
}
