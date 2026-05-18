'use client';

import { Button } from '@notopia-uit/ui/components/shadcn/button';
import { ArrowLeft, RefreshCcw } from 'lucide-react';
import { useRouter } from 'next/navigation';
import { useEffect } from 'react';

const ErrorPage = ({ error, reset }: { error: Error & { digest?: string }; reset: () => void }) => {
  const router = useRouter();

  useEffect(() => {
    console.error('Runtime Error:', error);
  }, [error]);

  return (
    <div className="grid min-h-screen w-full xl:grid-cols-2">
      <div className="flex flex-1 flex-col items-center justify-center p-8 text-center xl:items-start xl:text-start">
        <div className="mb-3 flex items-center gap-3">
          <span className="rounded-sm bg-red-100 px-2 py-1 font-mono text-xs font-semibold text-red-600">
            Error ID: {error.digest || 'Runtime_Error'}
          </span>
        </div>

        <h1 className="mb-2 text-4xl font-bold">Something Went Wrong</h1>

        <p className="text-muted-foreground">
          {error.message || 'Oops! Something went wrong on our end.'}
        </p>

        <div className="mt-8 flex flex-col gap-3 sm:flex-row">
          <Button onClick={() => reset()} variant="default" className="h-9 px-4 py-2">
            <RefreshCcw className="mr-2 size-4" />
            Try Again
          </Button>

          {/* Action 2: Go back home */}
          <Button onClick={() => router.push('/')} variant="outline" className="h-9 px-4 py-2">
            <ArrowLeft className="mr-2 size-4" />
            Go Back Home
          </Button>
        </div>
      </div>

      <div className="relative hidden xl:block">
        <img
          src="https://ui.shadcn.com/placeholder.svg"
          alt="Error illustration"
          className="absolute inset-0 size-full object-cover dark:brightness-[0.95] dark:invert"
        />
      </div>
    </div>
  );
};

export default ErrorPage;
