import { BillingBoundary } from './billing-boundary';

export default async function BillingPage({
  params,
  searchParams,
}: {
  params: Promise<{ workspaceId: string }>;
  searchParams: Promise<{ checkout?: string; portal?: string }>;
}) {
  const { workspaceId } = await params;
  const query = await searchParams;
  const returnType = ['success', 'cancel', 'payer'].includes(query.checkout ?? '')
    ? query.checkout!
    : query.portal === 'return'
      ? 'portal'
      : null;
  return <BillingBoundary workspaceId={workspaceId} returnType={returnType} />;
}
