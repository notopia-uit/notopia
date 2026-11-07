import { HocuspocusProviderWebsocketComponent, HocuspocusRoom } from '@hocuspocus/provider-react';
import { fetchAccessTokenServerSide } from '@lib/get-access-token';

import { EditorBoundary } from './editor-boundary';

export default async function NotePage({
  params,
}: {
  params: Promise<{ noteId: string; workspaceId: string }>;
}) {
  const { noteId, workspaceId } = await params;
  const token = await fetchAccessTokenServerSide();
  return (
    <div className="p-4">
      <HocuspocusProviderWebsocketComponent url={process.env.DOCUMENT_WS_URL}>
        <HocuspocusRoom name={noteId} token={token}>
          <EditorBoundary noteId={noteId} workspaceId={workspaceId} />
        </HocuspocusRoom>
      </HocuspocusProviderWebsocketComponent>
    </div>
  );
}
