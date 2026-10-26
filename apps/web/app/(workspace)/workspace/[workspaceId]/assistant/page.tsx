import { AssistantBoundary } from './assistant-boundary';

interface AssistantPageProps {
  params: Promise<{ workspaceId: string }>;
}

export default async function AssistantPage({ params }: AssistantPageProps) {
  const { workspaceId } = await params;
  return <AssistantBoundary workspaceId={workspaceId} />;
}
