'use client';

import { useQueryClient } from '@tanstack/react-query';
import { useCallback } from 'react';

interface UseQueryErrorHandlerProps {
  queryKey?: any[];
}

export function useQueryErrorHandler({ queryKey }: UseQueryErrorHandlerProps = {}) {
  const queryClient = useQueryClient();

  const retry = useCallback(() => {
    if (queryKey) {
      queryClient.invalidateQueries({ queryKey });
    } else {
      queryClient.invalidateQueries();
    }
  }, [queryClient, queryKey]);

  return { retry };
}
