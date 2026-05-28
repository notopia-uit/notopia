'use client';

import { useCallback, useRef, useEffect, useState } from 'react';

export function useSearchCache<T>(
  searchFn: (query: string) => Promise<T>,
  debounceMs: number = 300
) {
  const timeoutRef = useRef<NodeJS.Timeout | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<Error | null>(null);
  const inflightQueryRef = useRef<string | null>(null);

  useEffect(() => {
    return () => {
      if (timeoutRef.current) {
        clearTimeout(timeoutRef.current);
      }
    };
  }, []);

  const doFetch = useCallback(
    async (query: string): Promise<{ data: T; isLoading: boolean; error: Error | null }> => {
      setIsLoading(true);
      setError(null);

      try {
        const result = await searchFn(query);
        setIsLoading(false);
        return { data: result, isLoading: false, error: null };
      } catch (err) {
        const error = err instanceof Error ? err : new Error('Unknown error occurred');
        setError(error);
        setIsLoading(false);
        return { data: undefined as T, isLoading: false, error };
      }
    },
    [searchFn]
  );

  const search = useCallback(
    async (query: string): Promise<{ data: T; isLoading: boolean; error: Error | null }> => {
      if (timeoutRef.current) {
        clearTimeout(timeoutRef.current);
      }

      inflightQueryRef.current = query;

      return new Promise((resolve) => {
        timeoutRef.current = setTimeout(async () => {
          if (inflightQueryRef.current !== query) return;

          const result = await doFetch(query);
          resolve(result);
        }, debounceMs);
      });
    },
    [doFetch, debounceMs]
  );

  const clearCache = useCallback(() => {}, []);

  return { search, isLoading, error, clearCache };
}
