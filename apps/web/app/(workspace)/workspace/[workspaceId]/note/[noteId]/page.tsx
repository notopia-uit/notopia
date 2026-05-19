import { HocuspocusProviderWebsocketComponent, HocuspocusRoom } from '@hocuspocus/provider-react';
import { Editor } from '@ui/components/dynamic-editor';
import { fetchAccessTokenServerSide } from '@ui/lib/get-access-token';

export default async function NotePage({
  params,
}: {
  params: Promise<{ noteId: string; workspaceId: string }>;
}) {
  const { noteId, workspaceId } = await params;
  const token = await fetchAccessTokenServerSide();
  return (
    <div className="p-4">
      <HocuspocusProviderWebsocketComponent
        url={`ws://${process.env.NEXT_PUBLIC_API_URL}/document/ws/document`}
      >
        <HocuspocusRoom name={noteId} token={token}>
          <Editor noteId={noteId} workspaceId={workspaceId} />
        </HocuspocusRoom>
      </HocuspocusProviderWebsocketComponent>
    </div>
  );
}
