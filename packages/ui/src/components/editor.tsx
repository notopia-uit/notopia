"use client";

import "@blocknote/core/fonts/inter.css";
import { SuggestionMenuController, useCreateBlockNote } from "@blocknote/react";
import { BlockNoteView } from "@blocknote/shadcn";

import "@blocknote/shadcn/style.css";
import {
  createBlockNoteSchema,
  getNoteMenuItems,
  getTagMenuItems,
} from "@notopia-uit/ui/block-note";
import { useMemo, useState } from "react";

import { Icons } from "./icons";
import { Button } from "./shadcn/button";

export default function Editor({ noteId }: { noteId: string }) {
  // const { data: note } = useSuspenseQuery(
  //   getNoteOptions({
  //     path: {
  //       noteId: noteId,
  //     },
  //   })
  // );
  const mySchema = useMemo(() => createBlockNoteSchema(), []);

  const [isDirty, setIsDirty] = useState(false);
  const editor = useCreateBlockNote({
    schema: mySchema,
  });

  return (
    <div className="relative min-h-screen">
      <BlockNoteView
        editor={editor}
        onChange={() => {
          setIsDirty(true);
        }}
      >
        <SuggestionMenuController
          triggerCharacter={"#"}
          getItems={async (query) => {
            return Promise.resolve(getTagMenuItems(editor, query, []));
          }}
        />

        <SuggestionMenuController
          triggerCharacter={"@"}
          getItems={async (query) => {
            return Promise.resolve(getNoteMenuItems(editor, query, []));
          }}
        />
      </BlockNoteView>
      {isDirty && (
        <div className="animate-in fade-in slide-in-from-bottom-4 fixed bottom-10 left-1/2 -translate-x-1/2 duration-300">
          <Button
            variant="outline"
            size="icon"
            aria-label="save"
            // onClick={async () => {
            //   await saveContent(editor.document);
            // }}
            //
          >
            <Icons.Save />
          </Button>
        </div>
      )}
    </div>
  );
}
