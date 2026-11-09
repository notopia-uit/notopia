import { AssistantBoundary } from './assistant-boundary';

interface AssistantPageProps {
  params: Promise<{ workspaceId: string }>;
}

export default async function AssistantPage({ params }: AssistantPageProps) {
  const { workspaceId } = await params;
  const apiUrl = process.env.API_URL;
  if (!apiUrl) throw new Error('API_URL is not set');
  const origin = new URL(apiUrl.includes('://') ? apiUrl : `http://${apiUrl}`);
  origin.protocol = origin.protocol === 'https:' ? 'wss:' : 'ws:';
  origin.pathname = '/document/ws/document';
  return (
    <AssistantBoundary
      workspaceId={workspaceId}
      apiUrl={apiUrl}
      documentWsUrl={origin.toString()}
    />
  );
}
