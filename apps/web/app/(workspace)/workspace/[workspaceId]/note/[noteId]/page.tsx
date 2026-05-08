import { HocuspocusProviderWebsocketComponent, HocuspocusRoom } from '@hocuspocus/provider-react';
import { Editor } from '@ui/components/dynamic-editor';
import { fetchAccessTokenServerSide } from '@ui/lib/get-access-token';
export default async function NotePage({ params }: { params: Promise<{ noteId: string }> }) {
  const { noteId } = await params;
  return (
    <div className="p-4">
      <HocuspocusProviderWebsocketComponent url="ws://127.0.0.1:1234">
        <HocuspocusRoom name={noteId} token={async () => fetchAccessTokenServerSide()}>
          <Editor noteId={noteId} />
        </HocuspocusRoom>
      </HocuspocusProviderWebsocketComponent>
    </div>
  );
}
