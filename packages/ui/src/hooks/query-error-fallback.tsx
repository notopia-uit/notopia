'use client';

import { AlertCircleIcon, RefreshCcw } from 'lucide-react';
import { Button } from '../components/shadcn/button';
import { Alert, AlertDescription, AlertTitle } from '../components/shadcn/alert';

interface QueryErrorFallbackProps {
  error: unknown;
  onRetry?: () => void;
  title?: string;
  description?: string;
  compact?: boolean;
}

export function QueryErrorFallback({
  error,
  onRetry,
  title = 'Failed to Load',
  description,
  compact = false,
}: QueryErrorFallbackProps) {
  const getErrorMessage = (err: unknown): string => {
    if (err instanceof Error) {
      return err.message;
    }
    if (typeof err === 'object' && err !== null && 'message' in err) {
      return String((err as any).message);
    }
    return 'An unexpected error occurred';
  };

  const errorMessage = getErrorMessage(error);
  const fullDescription = description || errorMessage;

  if (compact) {
    return (
      <div className="flex items-center justify-between rounded-lg border border-red-200 bg-red-50 p-3 dark:border-red-900 dark:bg-red-950">
        <div className="flex items-center gap-2">
          <AlertCircleIcon className="size-4 text-red-600 dark:text-red-400" />
          <div className="text-sm">
            <p className="font-medium text-red-600 dark:text-red-400">{title}</p>
            <p className="text-xs text-red-500 dark:text-red-300">{fullDescription}</p>
          </div>
        </div>
        {onRetry && (
          <Button size="sm" variant="outline" onClick={onRetry} className="ml-2">
            <RefreshCcw className="size-3" />
          </Button>
        )}
      </div>
    );
  }

  return (
    <Alert variant="destructive">
      <AlertCircleIcon className="size-4" />
      <AlertTitle>{title}</AlertTitle>
      <AlertDescription className="mt-2 space-y-3">
        <p>{fullDescription}</p>
        {onRetry && (
          <Button onClick={onRetry} variant="outline" size="sm" className="mt-2">
            <RefreshCcw className="mr-2 size-4" />
            Try Again
          </Button>
        )}
      </AlertDescription>
    </Alert>
  );
}
