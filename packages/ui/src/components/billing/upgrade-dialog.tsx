'use client';

import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
} from '../shadcn/dialog';
import { BillingSettings } from './billing-settings';

export function UpgradeDialog({
  workspaceId,
  open,
  onOpenChange,
}: {
  workspaceId: string;
  open: boolean;
  onOpenChange: (open: boolean) => void;
}) {
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-h-[90dvh] overflow-y-auto sm:max-w-5xl">
        <DialogHeader>
          <DialogTitle>Upgrade your workspace</DialogTitle>
          <DialogDescription>More collaboration, history, and AI for your ideas.</DialogDescription>
        </DialogHeader>
        {open && <BillingSettings key={workspaceId} workspaceId={workspaceId} comparisonOnly />}
      </DialogContent>
    </Dialog>
  );
}
