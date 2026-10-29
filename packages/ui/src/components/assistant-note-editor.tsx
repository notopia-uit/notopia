'use client';

import '@notopia-uit/lib/yjs';
import '@notopia-uit/lib/hocuspocus';
import '@blocknote/core/fonts/inter.css';
import '@blocknote/shadcn/style.css';
import type { Chat } from '@ai-sdk/react';
import type { MyEditor } from '@blocknote/core';
import {
  HocuspocusProviderWebsocketComponent,
  HocuspocusRoom,
  useHocuspocusProvider,
  useHocuspocusSyncStatus,
  useHocuspocusConnectionStatus,
} from '@hocuspocus/provider-react';
import { useGetNoteQuery } from '@notopia-uit/api-gen';
import { useQuery } from '@tanstack/react-query';
import type { UIMessage } from 'ai';
import { useCallback, useEffect, useRef, useState } from 'react';

import { NavigationProvider } from '../contexts/navigation-context';
import { useEditorState } from '../hooks/use-editor-state';
import type { AssistantEditorSession, AssistantTarget } from '../lib/assistant-session';
import { getAuthClient } from '../lib/auth-client';
import { fetchAccessTokenClientSide } from '../lib/get-access-token-client-side';
import { EditorCore } from './editor-core';
import { Button } from './shadcn/button';
import { Spinner } from './shadcn/spinner';

interface Props {
  workspaceId: string;
  apiUrl: string;
  documentWsUrl: string;
  target: AssistantTarget;
  session: AssistantEditorSession<MyEditor>;
  chat: Chat<UIMessage>;
  busy: boolean;
  review: boolean;
  onAccept(): void;
  onReject(): void;
  onClose(): void;
  onError(error: Error): void;
}

function ConnectedEditor(props: Props) {
  const { target, session, chat, busy, review, onError } = props;
  const provider = useHocuspocusProvider();
  const sync = useHocuspocusSyncStatus();
  const connection = useHocuspocusConnectionStatus();
  const [ready, setReady] = useState(provider.synced);
  const editorRef = useRef<MyEditor | null>(null);
  const { data: identity } = getAuthClient().useSession();
  const { data: note } = useGetNoteQuery({ path: { noteId: target.noteId } });
  const { isModified, isCommitingDocument, handleSave } = useEditorState(target.noteId);
  useEffect(() => {
    const synced = () => {
      if (provider.synced) {
        setReady(true);
        if (editorRef.current) session.ready(target.sessionId, editorRef.current);
      }
    };
    provider.on('synced', synced);
    synced();
    return () => {
      provider.off('synced', synced);
    };
  }, [provider, session, target.sessionId]);
  const onReady = useCallback(
    (editor: MyEditor) => {
      editorRef.current = editor;
      if (provider.synced) session.ready(target.sessionId, editor);
    },
    [provider, session, target.sessionId]
  );
  useEffect(() => {
    if (connection !== 'connected') {
      session.unready(target.sessionId);
      if (ready && busy)
        onError(new Error('The live editor disconnected. Reconnect before editing.'));
    } else if (provider.synced && editorRef.current)
      session.ready(target.sessionId, editorRef.current);
  }, [ready, connection, busy, onError, provider, session, target.sessionId]);

  return (
    <section className="flex h-full min-h-0 flex-col" aria-label="Assistant target note">
      <header className="flex flex-wrap items-center gap-2 border-b p-3">
        <h2 className="min-w-0 flex-1 truncate font-medium">{note?.name ?? 'Opening note…'}</h2>
        <Button variant="ghost" disabled={busy || review} onClick={props.onClose}>
          Close
        </Button>
      </header>
      {review && (
        <div className="flex flex-wrap items-center gap-2 border-b p-3">
          <span className="mr-auto text-sm">Review the suggested changes</span>
          <Button onClick={props.onAccept}>Accept changes</Button>
          <Button variant="outline" onClick={props.onReject}>
            Discard changes
          </Button>
        </div>
      )}
      <div className="min-h-0 flex-1 overflow-auto">
        {ready ? (
          <NavigationProvider
            workspaceId={props.workspaceId}
            onNavigate={(href) => {
              window.location.assign(href);
            }}
          >
            <EditorCore
              noteId={target.noteId}
              apiUrl={props.apiUrl}
              assistantChat={chat}
              onEditorReady={onReady}
              sessionUser={{
                name: identity?.user?.name ?? 'User',
                color: 'var(--primary)',
                avatar: identity?.user?.image ?? '',
              }}
            />
          </NavigationProvider>
        ) : (
          <output className="flex items-center gap-2 p-4">
            <Spinner /> Connecting and syncing note…
          </output>
        )}
      </div>
      <footer className="flex items-center justify-between gap-2 border-t p-3">
        <span className="text-muted-foreground text-xs">
          Accept applies edits. Commit updates search.
        </span>
        <Button
          variant="outline"
          onClick={handleSave}
          disabled={
            !ready ||
            busy ||
            review ||
            connection !== 'connected' ||
            sync !== 'synced' ||
            !isModified ||
            isCommitingDocument
          }
        >
          {isCommitingDocument ? 'Committing…' : 'Commit'}
        </Button>
      </footer>
    </section>
  );
}

export function AssistantNoteEditor(props: Props) {
  const { session, target, onError } = props;
  const { data: token, error } = useQuery({
    queryKey: ['accessToken'],
    queryFn: fetchAccessTokenClientSide,
  });
  useEffect(() => {
    if (error) {
      session.fail(target.sessionId, error);
      onError(error);
    }
  }, [error, session, target.sessionId, onError]);
  if (!token)
    return (
      <output className="flex items-center gap-2 p-4">
        <Spinner /> Authorizing editor…
      </output>
    );
  return (
    <HocuspocusProviderWebsocketComponent url={props.documentWsUrl}>
      <HocuspocusRoom
        name={props.target.noteId}
        token={token}
        onAuthenticationFailed={() => {
          const failure = new Error('You no longer have access to this note.');
          props.session.fail(props.target.sessionId, failure);
          props.onError(failure);
        }}
      >
        <ConnectedEditor {...props} />
      </HocuspocusRoom>
    </HocuspocusProviderWebsocketComponent>
  );
}
