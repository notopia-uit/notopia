'use client';

import { Button } from '@notopia-uit/ui/components/shadcn/button';
import { ArrowLeft, RefreshCcw } from 'lucide-react';
import Link from 'next/link';
import { useEffect } from 'react';

export default function WorkspaceDetailError({
  error,
  reset,
}: {
  error: Error & { digest?: string };
  reset: () => void;
}) {
  useEffect(() => {
    console.error('Workspace detail error:', error);
  }, [error]);

  return (
    <div className="flex min-h-[50vh] flex-col items-center justify-center gap-4 p-8 text-center">
      <span className="rounded-sm bg-red-100 px-2 py-1 font-mono text-xs font-semibold text-red-600">
        {error.digest || 'Error'}
      </span>
      <h1 className="text-4xl font-bold">Workspace Error</h1>
      <p className="text-muted-foreground max-w-md">
        {error.message || 'Failed to load workspace.'}
      </p>
      <div className="mt-4 flex flex-col gap-3 sm:flex-row">
        <Button onClick={() => reset()} variant="default" size="sm">
          <RefreshCcw className="mr-2 size-4" />
          Try Again
        </Button>
        <Button asChild variant="outline" size="sm">
          <Link href="/workspace">
            <ArrowLeft className="mr-2 size-4" />
            Back to Workspaces
          </Link>
        </Button>
      </div>
    </div>
  );
}
