'use client';

import { previewCancelNowOptions, useCancelNowMutation } from '@notopia-uit/api-gen';
import { useQuery } from '@tanstack/react-query';
import { useState } from 'react';

import { billingError, formatMoney } from '../../lib/billing';
import { Button } from '../shadcn/button';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
} from '../shadcn/dialog';

export function CancelSubscriptionDialog({
  workspaceId,
  onCancelled,
  onClose,
}: {
  workspaceId: string;
  onCancelled: () => Promise<void>;
  onClose: () => void;
}) {
  const path = { workspaceId };
  const quote = useQuery({ ...previewCancelNowOptions({ path }), staleTime: 0, retry: false });
  const cancel = useCancelNowMutation();
  const [error, setError] = useState<string | null>(null);
  async function confirm() {
    setError(null);
    try {
      await cancel.mutateAsync({ path });
      await onCancelled();
      onClose();
    } catch (cause) {
      setError(billingError(cause));
    }
  }
  return (
    <Dialog
      open
      onOpenChange={(open) => {
        if (!open && !cancel.isPending) onClose();
      }}
    >
      <DialogContent showCloseButton={!cancel.isPending}>
        <DialogHeader>
          <DialogTitle>Cancel this subscription now?</DialogTitle>
          <DialogDescription>
            Paid access ends immediately for everyone in this workspace. To cancel at the end of
            your billing period instead, use Manage billing.
          </DialogDescription>
        </DialogHeader>
        {quote.isPending ? (
          <p>Loading your refund estimate…</p>
        ) : quote.error ? (
          <p className="text-destructive">{billingError(quote.error)}</p>
        ) : quote.data ? (
          <div className="bg-muted rounded-lg p-4">
            <p className="text-muted-foreground text-sm">Estimated refund</p>
            <p className="text-2xl font-semibold">
              {formatMoney(quote.data.amount, quote.data.currency)}
            </p>
            <p className="text-muted-foreground mt-2 text-sm">
              The final amount is recalculated when you confirm. Refund processing is asynchronous.
            </p>
          </div>
        ) : null}
        {error && (
          <p role="alert" className="text-destructive text-sm">
            {error}
          </p>
        )}
        <DialogFooter>
          <Button variant="outline" disabled={cancel.isPending} onClick={onClose}>
            Keep subscription
          </Button>
          <Button
            variant="destructive"
            disabled={!quote.data || quote.isFetching || cancel.isPending}
            onClick={() => void confirm()}
          >
            {cancel.isPending ? 'Canceling…' : 'Cancel now'}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
