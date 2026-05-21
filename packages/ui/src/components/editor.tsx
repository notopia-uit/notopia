'use client';

import '@notopia-uit/lib/yjs';
import '@notopia-uit/lib/hocuspocus';
import '@blocknote/core/fonts/inter.css';
import '@blocknote/shadcn/style.css';
import { Spinner } from '@notopia-uit/ui/components/shadcn/spinner';
import { useEditorState } from '@notopia-uit/ui/hooks/use-editor-state';
import { authClient } from '@notopia-uit/ui/lib/auth-client';
import { useRouter } from 'next/navigation';
import { useState, useEffect, useMemo } from 'react';
import { useHocuspocusProvider } from '@hocuspocus/provider-react';

import { getDeterministicColor } from './../lib/utils/color';
import { EditorCore } from './editor-core';
import { EditorToolbar } from './editor-toolbar';
import { ErrorAlert } from './error-alert';
import { Icons } from './icons';
import { NoteTitle } from './note-title';
import { Button } from './shadcn/button';
import { SuccessAlert } from './success-alert';

export default function Editor({ noteId, workspaceId }: { noteId: string; workspaceId?: string }) {
  const { data: sessionData } = authClient.useSession();
  const router = useRouter();
  const [isAwarenessReady, setIsAwarenessReady] = useState(false);
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

  const { isModified, isCommitingDocument, alert, handleSave } = useEditorState(noteId);

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
      <EditorToolbar
        noteId={noteId}
        currentEditor={undefined}
      />
      {isAwarenessReady ? (
        <EditorCore sessionUser={sessionUser} />
      ) : (
        <div className="flex items-center justify-center h-96">
          <Spinner />
        </div>
      )}

      {alert?.type === 'success' && <SuccessAlert title={alert.title} message={alert.message} />}
      {alert?.type === 'error' && <ErrorAlert title={alert.title} message={alert.message} />}

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
