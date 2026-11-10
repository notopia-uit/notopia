'use client';

import { BillingSettings } from '@notopia-uit/ui/components/billing/billing-settings';
import { ErrorBoundary } from '@notopia-uit/ui/components/error-boundary';

export function BillingBoundary({
  workspaceId,
  returnType,
}: {
  workspaceId: string;
  returnType: string | null;
}) {
  return (
    <ErrorBoundary fallbackTitle="Billing could not be loaded">
      <BillingSettings
        key={`${workspaceId}:${returnType}`}
        workspaceId={workspaceId}
        returnType={returnType}
      />
    </ErrorBoundary>
  );
}
