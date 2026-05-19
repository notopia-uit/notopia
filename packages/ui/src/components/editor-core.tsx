'use client';

import { SuggestionMenuController, useCreateBlockNote } from '@blocknote/react';
import { BlockNoteView } from '@blocknote/shadcn';
import { BlockNoteEditor } from '@blocknote/core';
import { useHocuspocusProvider } from '@hocuspocus/provider-react';
import {
  createBlockNoteSchema,
  getNoteMenuItems,
  getTagMenuItems,
} from '@notopia-uit/ui/block-note';
import { ForwardedRef, forwardRef, useMemo } from 'react';

interface EditorCoreProps {
  sessionUser?: {
    name: string;
    color: string;
    avatar: string;
  };
}

export const EditorCore = forwardRef<BlockNoteEditor | null, EditorCoreProps>(
  function EditorCore({ sessionUser }, ref) {
    const mySchema = useMemo(() => createBlockNoteSchema(), []);
    const provider = useHocuspocusProvider();

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

    return (
      <BlockNoteView editor={editor}>
        <SuggestionMenuController
          triggerCharacter={'#'}
          getItems={async (query) => {
            return Promise.resolve(getTagMenuItems(editor, query, []));
          }}
        />

        <SuggestionMenuController
          triggerCharacter={'@'}
          getItems={async (query) => {
            return Promise.resolve(getNoteMenuItems(editor, query, []));
          }}
        />
      </BlockNoteView>
    );
  }
);


