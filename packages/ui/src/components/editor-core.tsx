'use client';

import { BlockNoteEditor } from '@blocknote/core';
import { SuggestionMenuController, useCreateBlockNote } from '@blocknote/react';
import { BlockNoteView } from '@blocknote/shadcn';
import {
  useHocuspocusProvider,
  useHocuspocusConnectionStatus,
} from '@hocuspocus/provider-react';
import {
  createBlockNoteSchema,
  getNoteMenuItems,
  getTagMenuItems,
} from '@notopia-uit/ui/block-note';
import { forwardRef, useMemo, useEffect, useState } from 'react';
import { Spinner } from '@notopia-uit/ui/components/shadcn/spinner';

interface EditorCoreProps {
  sessionUser?: {
    name: string;
    color: string;
    avatar: string;
  };
}

export const EditorCore = forwardRef<BlockNoteEditor | null, EditorCoreProps>(function EditorCore(
  { sessionUser },
  _ref
) {
  const mySchema = useMemo(() => createBlockNoteSchema(), []);
  const provider = useHocuspocusProvider();
  const connectionStatus = useHocuspocusConnectionStatus();
  const [isReady, setIsReady] = useState(false);

  useEffect(() => {
    setIsReady(connectionStatus === 'connected');
  }, [connectionStatus]);

  const editor = useCreateBlockNote({
    schema: mySchema,
    collaboration: {
      provider: {
        awareness: provider.awareness ? provider.awareness : undefined,
      },
      fragment: provider.document.getXmlFragment('prosemirror'),
      user: sessionUser || {
        name: 'Anonymous',
        color: '#999999',
      },
    },
  });

  if (!isReady) {
    return (
      <div className="flex items-center justify-center h-96">
        <Spinner />
      </div>
    );
  }

  return (
    <BlockNoteView editor={editor}>
      <SuggestionMenuController
        triggerCharacter={'#'}
        getItems={async (query) => {
          return Promise.resolve(getTagMenuItems(editor, query, []));
        }}
      />

      <SuggestionMenuController
        triggerCharacter={'[['}
        getItems={async (query) => {
          return Promise.resolve(getNoteMenuItems(editor, query, []));
        }}
      />
    </BlockNoteView>
  );
});
