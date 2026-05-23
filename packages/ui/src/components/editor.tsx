'use client';

import '@notopia-uit/lib/yjs';
import '@notopia-uit/lib/hocuspocus';
import '@blocknote/core/fonts/inter.css';
import '@blocknote/shadcn/style.css';
import { useHocuspocusProvider } from '@hocuspocus/provider-react';
import { EditorSearchProvider } from '@notopia-uit/ui/components/editor-search-provider';
import { Spinner } from '@notopia-uit/ui/components/shadcn/spinner';
import { useEditorState } from '@notopia-uit/ui/hooks/use-editor-state';
import { authClient } from '@notopia-uit/ui/lib/auth-client';
import { useRouter } from 'next/navigation';
import { useState, useEffect, useMemo, useRef } from 'react';

import { getDeterministicColor } from './../lib/utils/color';
import { EditorCore } from './editor-core';
import { EditorToolbar } from './editor-toolbar';
import { Icons } from './icons';
import { NoteTitle } from './note-title';
import { Button } from './shadcn/button';

export default function Editor({ noteId, workspaceId }: { noteId: string; workspaceId?: string }) {
  const { data: sessionData } = authClient.useSession();
  const router = useRouter();
  const [isAwarenessReady, setIsAwarenessReady] = useState(true);
  const provider = useHocuspocusProvider();

  const sessionUser = useMemo(
    () => ({
      name: sessionData?.user?.name ?? 'Anonymous',
      color: getDeterministicColor(sessionData?.user?.id ?? 'anonymous'),
      avatar: sessionData?.user?.image ?? 'https://placehold.net/default.svg',
    }),
    [sessionData?.user?.name, sessionData?.user?.id, sessionData?.user?.image]
  );

  useEffect(() => {
    if (provider.awareness && sessionUser) {
      provider.awareness.setLocalState(sessionUser);
      setIsAwarenessReady(true);
    }
  }, [provider.awareness, sessionUser]);

  const editorRef = useRef<any>(null);
  const { isModified, isCommitingDocument, handleSave } = useEditorState(noteId);

  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if ((e.ctrlKey || e.metaKey) && e.key === 'g') {
        e.preventDefault();
        router.push(`/workspace/${workspaceId}/note/${noteId}/graph`);
      }
    };

    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [router, workspaceId, noteId]);

  return (
    <div className="relative min-h-screen">
      <NoteTitle noteId={noteId} workspaceId={workspaceId} />
      <EditorToolbar noteId={noteId} currentEditor={editorRef.current} />
      {isAwarenessReady ? (
        <EditorSearchProvider
          workspaceId={workspaceId || ''}
          host={process.env.NEXT_PUBLIC_MEILISEARCH_HOST || 'http://localhost:7700'}
        >
          <EditorCore ref={editorRef} sessionUser={sessionUser} noteId={noteId} />
        </EditorSearchProvider>
      ) : (
        <div className="flex h-96 items-center justify-center">
          <Spinner />
        </div>
      )}

      {isModified && (
        <div className="animate-in fade-in slide-in-from-bottom-4 fixed bottom-10 left-1/2 -translate-x-1/2 duration-300">
          {isCommitingDocument ? (
            <Spinner />
          ) : (
            <Button variant="outline" size="icon" aria-label="save" onClick={handleSave}>
              <Icons.Save />
            </Button>
          )}
        </div>
      )}
    </div>
  );
}
