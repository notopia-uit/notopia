import { HocuspocusProviderWebsocketComponent, HocuspocusRoom } from '@hocuspocus/provider-react';
import { Editor } from '@ui/components/dynamic-editor';
import { fetchAccessTokenServerSide } from '@ui/lib/get-access-token';
export default async function NotePage({ params }: { params: Promise<{ noteId: string }> }) {
  const { noteId } = await params;
  const token = await fetchAccessTokenServerSide();
  return (
    <div className="p-4">
      <HocuspocusProviderWebsocketComponent url= {process.env.NEXT_PUBLIC_HOCUSPOCUS_SERVER_URL ?? 'ws://localhost:4000'} >
        <HocuspocusRoom name={noteId} token={token}>
          <Editor noteId={noteId} />
        </HocuspocusRoom>
      </HocuspocusProviderWebsocketComponent>
    </div>
  );
}
