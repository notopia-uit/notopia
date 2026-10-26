'use client';

import '@notopia-uit/lib/yjs';
import '@notopia-uit/lib/hocuspocus';
import '@blocknote/core/fonts/inter.css';
import '@blocknote/shadcn/style.css';
import { getMyWorkspacesOptions } from '@notopia-uit/api-gen';
import { Spinner } from '@notopia-uit/ui/components/shadcn/spinner';
import { NavigationProvider } from '@notopia-uit/ui/contexts/navigation-context';
import { useEditorState } from '@notopia-uit/ui/hooks/use-editor-state';
import { getAuthClient } from '@notopia-uit/ui/lib/auth-client';
import { useQuery } from '@tanstack/react-query';
import { useEffect, useMemo, useRef, useState } from 'react';

import { getDeterministicColor } from './../lib/utils/color';
import { EditorCore } from './editor-core';
import { EditorToolbar } from './editor-toolbar';
import { Icons } from './icons';
import { NoteTitle } from './note-title';
import { Button } from './shadcn/button';
import { TableOfContents } from './table-of-contents';

export default function Editor({
  noteId,
  workspaceId,
  onNavigate,
  apiUrl,
}: {
  noteId: string;
  workspaceId?: string;
  onNavigate: (href: string) => void;
  apiUrl: string;
}) {
  const { data: sessionData } = getAuthClient().useSession();

  const sessionUser = useMemo(
    () => ({
      name: sessionData?.user?.name ?? 'Anonymous',
      color: getDeterministicColor(sessionData?.user?.id ?? 'anonymous'),
      avatar: sessionData?.user?.image ?? '',
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
        onNavigate(`/workspace/${workspaceId}/note/${noteId}/graph`);
      }
    };

    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [onNavigate, workspaceId, noteId]);

  return (
    <div className="relative min-h-screen">
      <NoteTitle noteId={noteId} workspaceId={workspaceId} />
      <EditorToolbar
        noteId={noteId}
        workspaceId={workspaceId ?? ''}
        currentEditor={editorRef.current}
        onNavigate={onNavigate}
      />
      <NavigationProvider workspaceId={workspaceId ?? ''} onNavigate={onNavigate}>
        <EditorCore
          ref={editorRef}
          sessionUser={sessionUser}
          noteId={noteId}
          isViewer={isViewer}
          apiUrl={apiUrl}
          onEditorReady={setEditorInstance}
        />
      </NavigationProvider>

      {editorInstance && <TableOfContents editor={editorInstance} />}

      {!isViewer && (isModified || isCommitingDocument) && (
        <div className="animate-in fade-in slide-in-from-bottom-4 fixed bottom-10 left-1/2 -translate-x-1/2 duration-300">
          <Button variant="outline" onClick={handleSave} disabled={isCommitingDocument}>
            {isCommitingDocument ? <Spinner /> : <Icons.Save />}
            {isCommitingDocument ? 'Committing…' : 'Commit changes'}
          </Button>
        </div>
      )}
    </div>
  );
}
