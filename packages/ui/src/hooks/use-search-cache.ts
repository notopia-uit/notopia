'use client';

import { useCallback, useRef, useEffect, useState } from 'react';

interface CacheEntry<T> {
  data: T;
  timestamp: number;
}

const CACHE_TTL = 5 * 60 * 1000; // 5 minutes

export function useSearchCache<T>(
  searchFn: (query: string) => Promise<T>,
  debounceMs: number = 300
) {
  const cacheRef = useRef<Map<string, CacheEntry<T>>>(new Map());
  const timeoutRef = useRef<NodeJS.Timeout | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    return () => {
      if (timeoutRef.current) {
        clearTimeout(timeoutRef.current);
      }
    };
  }, []);

  const search = useCallback(
    async (query: string): Promise<{ data: T; isLoading: boolean; error: Error | null }> => {
      if (timeoutRef.current) {
        clearTimeout(timeoutRef.current);
      }

      setIsLoading(true);
      setError(null);

      return new Promise((resolve) => {
        timeoutRef.current = setTimeout(async () => {
          try {
            const result = await searchFn(query);
            setIsLoading(false);
            setError(null);
            resolve({ data: result, isLoading: false, error: null });
          } catch (err) {
            const error = err instanceof Error ? err : new Error('Unknown error occurred');
            setError(error);
            setIsLoading(false);
            resolve({ data: undefined as T, isLoading: false, error });
          }
        }, debounceMs);
      });
    },
    [searchFn, debounceMs]
  );

  const clearCache = useCallback(() => {
    cacheRef.current.clear();
  }, []);

  return { search, isLoading, error, clearCache };
}
