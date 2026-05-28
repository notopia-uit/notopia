'use client';

import '@notopia-uit/lib/yjs';
import '@notopia-uit/lib/hocuspocus';
import '@blocknote/core/fonts/inter.css';
import '@blocknote/shadcn/style.css';
import { Spinner } from '@notopia-uit/ui/components/shadcn/spinner';
import { useEditorState } from '@notopia-uit/ui/hooks/use-editor-state';
import { getAuthClient } from '@notopia-uit/ui/lib/auth-client';
import { getMyWorkspacesOptions } from '@notopia-uit/api-gen';
import { useQuery } from '@tanstack/react-query';
import { useRouter } from 'next/navigation';
import { useEffect, useMemo, useRef, useState } from 'react';

import { getDeterministicColor } from './../lib/utils/color';
import { EditorCore } from './editor-core';
import { EditorToolbar } from './editor-toolbar';
import { Icons } from './icons';
import { NoteTitle } from './note-title';
import { Button } from './shadcn/button';
import { TableOfContents } from './table-of-contents';

export default function Editor({ noteId, workspaceId }: { noteId: string; workspaceId?: string }) {
  const { data: sessionData } = getAuthClient().useSession();
  const router = useRouter();

  const sessionUser = useMemo(
    () => ({
      name: sessionData?.user?.name ?? 'Anonymous',
      color: getDeterministicColor(sessionData?.user?.id ?? 'anonymous'),
      avatar: sessionData?.user?.image ?? 'https://placehold.net/default.svg',
    }),
    [sessionData?.user?.name, sessionData?.user?.id, sessionData?.user?.image]
  );

  const editorRef = useRef<any>(null);
  const [editorInstance, setEditorInstance] = useState<any>(null);
  const { isModified, isCommitingDocument, handleSave } = useEditorState(noteId);

  const { data: allWorkspaceData } = useQuery({
    ...getMyWorkspacesOptions({}),
  });
  const currentWorkspace = allWorkspaceData?.find((ws) => ws.workspace.id === workspaceId);
  const isViewer = currentWorkspace?.role === 'viewer';

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
      <EditorCore
        ref={editorRef}
        sessionUser={sessionUser}
        noteId={noteId}
        isViewer={isViewer}
        onEditorReady={setEditorInstance}
      />

      {editorInstance && <TableOfContents editor={editorInstance} />}

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
