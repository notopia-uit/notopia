'use client';

import Link from 'next/link';

import { Button } from './shadcn/button';

export interface NotFoundProps {
  /** The resource type that was not found (e.g. "Note", "Workspace"). */
  resource?: string;
}

export default function NotFound({ resource = 'Page' }: NotFoundProps) {
  return (
    <div className="flex min-h-[50vh] flex-col items-center justify-center gap-4 p-8 text-center">
      <span className="rounded-sm bg-red-100 px-2 py-1 font-mono text-xs font-semibold text-red-600 dark:bg-red-900 dark:text-red-300">
        404
      </span>
      <h1 className="text-4xl font-bold">{resource} Not Found</h1>
      <p className="text-muted-foreground max-w-md">
        The {resource.toLowerCase()} you are looking for does not exist or has been removed.
      </p>
      <Button asChild variant="outline" className="mt-4">
        <Link href="/">Go Home</Link>
      </Button>
    </div>
  );
}
