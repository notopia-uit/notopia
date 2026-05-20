'use client';

import '@notopia-uit/lib/yjs';
import '@notopia-uit/lib/hocuspocus';
import '@blocknote/core/fonts/inter.css';
import '@blocknote/shadcn/style.css';

import { useState } from 'react';
import { authClient } from '@notopia-uit/ui/lib/auth-client';
import { getDeterministicColor } from './../lib/utils/color';
import { useEditorState } from '@notopia-uit/ui/hooks/use-editor-state';

import { EditorStatus } from './editor-status';
import { EditorToolbar } from './editor-toolbar';
import { EditorCore } from './editor-core';
import { NoteTitle } from './note-title';
import { ErrorAlert } from './error-alert';
import { SuccessAlert } from './success-alert';
import { NoteGraphModal } from './note-graph-modal';
import { Spinner } from '@notopia-uit/ui/components/shadcn/spinner';
import { Button } from './shadcn/button';
import { Icons } from './icons';

export default function Editor({ noteId, workspaceId }: { noteId: string; workspaceId?: string }) {
  const { data: sessionData } = authClient.useSession();
  const [isGraphModalOpen, setIsGraphModalOpen] = useState(false);
  
  const { isModified, isCommitingDocument, alert, handleSave } = useEditorState(noteId);

  const sessionUser = {
    name: sessionData?.user?.name ?? 'Anonymous',
    color: getDeterministicColor(sessionData?.user?.id ?? 'anonymous'),
    avatar: sessionData?.user?.image ?? 'https://placehold.net/default.svg',
  };

  return (
    <div className="relative min-h-screen">
      <EditorStatus />
      <NoteTitle 
        noteId={noteId}
        workspaceId={workspaceId}
      />
      <EditorToolbar 
        noteId={noteId} 
        currentEditor={undefined}
        onGraphOpen={() => setIsGraphModalOpen(true)}
      />
      <EditorCore sessionUser={sessionUser} />
      
      {alert?.type === 'success' && <SuccessAlert title={alert.title} message={alert.message} />}
      {alert?.type === 'error' && <ErrorAlert title={alert.title} message={alert.message} />}
      
      {isModified && (
        <div className="animate-in fade-in slide-in-from-bottom-4 fixed bottom-10 left-1/2 -translate-x-1/2 duration-300">
          {isCommitingDocument ? (
            <Spinner />
          ) : (
            <Button
              variant="outline"
              size="icon"
              aria-label="save"
              onClick={handleSave}
            >
              <Icons.Save />
            </Button>
          )}
        </div>
      )}
      <NoteGraphModal isOpen={isGraphModalOpen} onOpenChange={setIsGraphModalOpen} noteId={noteId} />
    </div>
  );
}
