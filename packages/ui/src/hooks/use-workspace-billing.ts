'use client';

import {
  getMyWorkspacesOptions,
  listPlansOptions,
  getWorkspaceEntitlementsOptions,
  getWorkspaceBillingOptions,
  useCreateCheckoutSessionMutation,
  useCreatePortalSessionMutation,
  useClaimPayerMutation,
} from '@notopia-uit/api-gen';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import { useCallback, useRef, useState } from 'react';

import { billingError } from '../lib/billing';

export function useWorkspaceBilling(workspaceId: string) {
  const client = useQueryClient();
  const membership = useQuery(getMyWorkspacesOptions());
  const workspace = membership.data?.find((item) => item.workspace.id === workspaceId);
  const isOwner = workspace?.role === 'owner';
  const path = { workspaceId };
  const plans = useQuery(listPlansOptions());
  const entitlements = useQuery({
    ...getWorkspaceEntitlementsOptions({ path }),
    enabled: !!workspace,
  });
  const billing = useQuery({
    ...getWorkspaceBillingOptions({ path }),
    enabled: isOwner,
    retry: false,
  });
  const checkout = useCreateCheckoutSessionMutation();
  const portal = useCreatePortalSessionMutation();
  const claim = useClaimPayerMutation();
  const lock = useRef(false);
  const [pending, setPending] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const refresh = useCallback(async () => {
    const path = { workspaceId };
    await Promise.all([
      client.invalidateQueries({ queryKey: getWorkspaceEntitlementsOptions({ path }).queryKey }),
      client.invalidateQueries({ queryKey: getWorkspaceBillingOptions({ path }).queryKey }),
      client.invalidateQueries({ queryKey: listPlansOptions().queryKey }),
    ]);
  }, [client, workspaceId]);

  async function redirect(action: 'checkout' | 'portal' | 'claim', priceId?: string) {
    if (!isOwner || lock.current) return;
    lock.current = true;
    setPending(true);
    setError(null);
    try {
      const result =
        action === 'checkout'
          ? await checkout.mutateAsync({ path, body: { priceId: priceId! } })
          : action === 'portal'
            ? await portal.mutateAsync({ path })
            : await claim.mutateAsync({ path });
      if (action === 'checkout') {
        const selected = plans.data
          ?.flatMap((plan) => plan.prices)
          .find((price) => price.id === priceId);
        if (selected) {
          try {
            sessionStorage.setItem(
              `billing-checkout:${workspaceId}`,
              JSON.stringify({
                plan: selected.plan,
                interval: selected.interval,
                previousId: billing.data?.subscription?.id ?? null,
              })
            );
          } catch (cause) {
            console.warn('Unable to save the local checkout reference', cause);
          }
        }
      }
      window.location.assign(result.url);
    } catch (cause) {
      setError(billingError(cause));
      await refresh();
    } finally {
      lock.current = false;
      setPending(false);
    }
  }

  return {
    workspace,
    isOwner,
    plans: plans.data ?? [],
    entitlements: entitlements.data,
    billing: isOwner ? billing.data : undefined,
    pending,
    error,
    redirect,
    refresh,
    loading:
      membership.isPending ||
      plans.isPending ||
      (!!workspace && entitlements.isPending) ||
      (isOwner && billing.isPending),
    loadError:
      membership.error ?? plans.error ?? entitlements.error ?? (isOwner ? billing.error : null),
    missingWorkspace: membership.isSuccess && !workspace,
  };
}

export type WorkspaceBillingState = ReturnType<typeof useWorkspaceBilling>;
