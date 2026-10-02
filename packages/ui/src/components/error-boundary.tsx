'use client';

import { Component, type ReactNode } from 'react';

import { Button } from './shadcn/button';

export interface ErrorBoundaryProps {
  children: ReactNode;
  fallbackTitle?: string;
  fallback?: (error: Error, retry: () => void) => ReactNode;
}

interface ErrorBoundaryState {
  error: Error | null;
}

export class ErrorBoundary extends Component<ErrorBoundaryProps, ErrorBoundaryState> {
  override state: ErrorBoundaryState = { error: null };

  static getDerivedStateFromError(error: Error): ErrorBoundaryState {
    return { error };
  }

  override componentDidCatch(error: Error, info: React.ErrorInfo) {
    console.error('[ErrorBoundary]', error, info.componentStack);
  }

  retry = () => {
    this.setState({ error: null });
  };

  override render() {
    const { error } = this.state;
    const { children, fallback, fallbackTitle } = this.props;

    if (error) {
      if (fallback) return fallback(error, this.retry);
      return <DefaultErrorFallback error={error} title={fallbackTitle} retry={this.retry} />;
    }
    return children;
  }
}

function DefaultErrorFallback({
  error,
  title = 'Something went wrong',
  retry,
}: {
  error: Error;
  title?: string;
  retry: () => void;
}) {
  return (
    <div className="flex flex-col items-center justify-center gap-4 p-8 text-center">
      <div className="rounded-sm bg-red-100 px-2 py-1 font-mono text-xs font-semibold text-red-600 dark:bg-red-900 dark:text-red-300">
        {error.name}
      </div>
      <h2 className="text-xl font-bold">{title}</h2>
      <p className="text-muted-foreground max-w-md text-sm">{error.message}</p>
      <Button onClick={retry} variant="outline" size="sm">
        Try again
      </Button>
    </div>
  );
}
