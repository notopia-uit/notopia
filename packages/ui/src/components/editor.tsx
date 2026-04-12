'use client';
import { BlockNoteEditor, PartialBlock } from '@blocknote/core';
import { insertOrUpdateBlockForSlashMenu } from '@blocknote/core/extensions';
import '@blocknote/core/fonts/inter.css';
import {
  DefaultReactSuggestionItem,
  SuggestionMenuController,
  getDefaultReactSlashMenuItems,
  useCreateBlockNote,
} from '@blocknote/react';
import { BlockNoteView } from '@blocknote/shadcn';
import '@blocknote/shadcn/style.css';
import { getNoteOptions } from '@notopia-uit/api-gen';
import { useSuspenseQuery } from '@tanstack/react-query';
import { useState } from 'react';

import { Icons } from './icons';
import { Button } from './ui/button';

const insertHelloWorldItem = (editor: BlockNoteEditor) => ({
  title: 'Insert Hello World',
  onItemClick: () =>
    insertOrUpdateBlockForSlashMenu(editor, {
      type: 'paragraph',
      content: [{ type: 'text', text: 'Hello World', styles: { bold: true } }],
    }),
  aliases: ['helloworld', 'hw'],
  group: 'Other',
  icon: <Icons.Logo />,
  subtext: "Used to insert a block with 'Hello World' below.",
});

const getCustomSlashMenuItems = (
  editor: BlockNoteEditor
): DefaultReactSuggestionItem[] => [
  ...getDefaultReactSlashMenuItems(editor),
  insertHelloWorldItem(editor),
];

interface EditorProps {
  initialContent: Promise<PartialBlock[]>;
}

export default function Editor({ noteId }: { noteId: string }) {
  // const { data: note } = useSuspenseQuery(
  //   getNoteOptions({
  //     path: {
  //       noteId: noteId,
  //     },
  //   })
  // );
  const [isDirty, setIsDirty] = useState(false);
  const editor = useCreateBlockNote({});

  return (
    <div className="relative min-h-screen">
      <BlockNoteView
        editor={editor}
        onChange={() => {
          setIsDirty(true);
        }}
      >
        <SuggestionMenuController
          triggerCharacter="/"
          // getItems={async (query) =>
          //   filterSuggestionItems(getCustomSlashMenuItems(editor), query)
          // }
        />
      </BlockNoteView>
      {isDirty && (
        <div className="fixed bottom-10 left-1/2 -translate-x-1/2 animate-in fade-in slide-in-from-bottom-4 duration-300">
          <Button
            variant="outline"
            size="icon"
            aria-label="save"
            // onClick={async () => {
            //   await saveContent(editor.document);
            // }}
          >
            <Icons.Save />
          </Button>
        </div>
      )}
    </div>
  );
}
export type { EditorProps };
