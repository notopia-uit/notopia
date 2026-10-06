import { AssistantBoundary } from './assistant-boundary';

interface AssistantPageProps {
  params: Promise<{ workspaceId: string }>;
}

export default async function AssistantPage({ params }: AssistantPageProps) {
  const { workspaceId } = await params;
  const apiUrl = process.env.API_URL;
  if (!apiUrl) {
    throw new Error('API_URL is not set');
  }
  return <AssistantBoundary workspaceId={workspaceId} aiApiUrl={`http://${apiUrl}/ai`} />;
}
