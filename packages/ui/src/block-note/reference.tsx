'use client';

import { createReactInlineContentSpec } from '@blocknote/react';
import { HocuspocusProviderWebsocketComponent, HocuspocusRoom } from '@hocuspocus/provider-react';
import { ReferenceConfig, ReferenceInlineContentSpec } from '@notopia-uit/lib/block-note';
import { useGetNoteQuery } from '@notopia-uit/api-gen';
import { EditorCore } from '@notopia-uit/ui/components/editor-core';
import { Dialog, DialogContent } from '@notopia-uit/ui/components/shadcn/dialog';
import { Spinner } from '@notopia-uit/ui/components/shadcn/spinner';
import { getAuthClient } from '@notopia-uit/ui/lib/auth-client';
import { fetchAccessTokenClientSide } from '@notopia-uit/ui/lib/get-access-token-client-side';
import { getDeterministicColor } from '@notopia-uit/ui/lib/utils/color';
import { useEffect, useMemo, useState } from 'react';

function ReferencePreview({
  noteId,
  open,
  onOpenChange,
  previewWsUrl,
}: {
  noteId: string;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  previewWsUrl: string;
}) {
  const [token, setToken] = useState<string | null>(null);
  const { data: sessionData } = getAuthClient().useSession();

  const sessionUser = useMemo(
    () => ({
      name: sessionData?.user?.name ?? 'Anonymous',
      color: getDeterministicColor(sessionData?.user?.id ?? 'anonymous'),
      avatar: sessionData?.user?.image ?? 'https://placehold.net/default.svg',
    }),
    [sessionData?.user?.name, sessionData?.user?.id, sessionData?.user?.image]
  );

  useEffect(() => {
    if (open && !token) {
      fetchAccessTokenClientSide().then(setToken).catch(console.error);
    }
  }, [open, token]);

  const handleOpenChange = (open: boolean) => {
    onOpenChange(open);
  };

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent
        className="flex max-h-[80vh] max-w-4xl flex-col gap-0 overflow-hidden p-0"
        showCloseButton={true}
      >
        <div className="min-h-75 flex-1 overflow-auto">
          {open && token ? (
            <HocuspocusProviderWebsocketComponent url={previewWsUrl}>
              <HocuspocusRoom name={noteId} token={token}>
                <EditorCore sessionUser={sessionUser} noteId={noteId} isViewer={true} />
              </HocuspocusRoom>
            </HocuspocusProviderWebsocketComponent>
          ) : (
            <div className="flex h-full min-h-75 items-center justify-center">
              <Spinner />
            </div>
          )}
        </div>
      </DialogContent>
    </Dialog>
  );
}

const ReferenceLink = ({
  noteId,
  previewWsUrl,
}: {
  noteId: string;
  previewWsUrl: string;
}) => {
  const { data: note, isPending, isError } = useGetNoteQuery({
    path: { noteId },
  });
  const [showPreview, setShowPreview] = useState(false);

  const displayName = isPending ? '' : isError ? 'Unknown Note' : note?.name || 'Untitled Note';

  return (
    <>
      <a
        href={`/note/${noteId}`}
        className="notopia-reference bg-primary/10 text-primary hover:bg-primary/20 cursor-pointer rounded-sm px-1"
        data-notopia-ref={noteId}
        onMouseEnter={() => setShowPreview(true)}
      >
        @{displayName}
      </a>
      <ReferencePreview
        noteId={noteId}
        open={showPreview}
        onOpenChange={setShowPreview}
        previewWsUrl={previewWsUrl}
      />
    </>
  );
};

export const createBlockNoteReferenceSpec = (apiUrl?: string): ReferenceInlineContentSpec => {
  const previewWsUrl = `ws://${apiUrl || 'api.notopia.localhost'}/document/ws/document`;

  return createReactInlineContentSpec(ReferenceConfig, {
    render: (props) => {
      return <ReferenceLink noteId={props.inlineContent.props.noteId} previewWsUrl={previewWsUrl} />;
    },
    toExternalHTML: (props) => {
      const id = props.inlineContent.props.noteId;
      return (
        <a href={`@${id}`} data-notopia-ref={id}>
          @{props.inlineContent.props.noteId}
        </a>
      );
    },

    parse: (element) => {
      const noteId = element.getAttribute('data-notopia-ref');
      if (!noteId) {
        return undefined;
      } else {
        return {
          noteId,
        };
      }
    },
  });
};
