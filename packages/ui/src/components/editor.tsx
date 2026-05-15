'use client';

// we have to import the prototype augment, type augment for lib/yjs and hocuspocus
// oxfmt-ignore
import '@notopia-uit/lib/yjs';
// oxfmt-ignore
import '@notopia-uit/lib/hocuspocus';

import '@blocknote/core/fonts/inter.css';
import { SuggestionMenuController, useCreateBlockNote } from '@blocknote/react';
import { BlockNoteView } from '@blocknote/shadcn';
import {
  useHocuspocusAwareness,
  useHocuspocusSyncStatus,
  useHocuspocusConnectionStatus,
  useHocuspocusProvider,
} from '@hocuspocus/provider-react';
import { useCommitDocumentMutation } from '@notopia-uit/api-gen';
import {
  createBlockNoteSchema,
  getNoteMenuItems,
  getTagMenuItems,
} from '@notopia-uit/ui/block-note';
import { Spinner } from '@notopia-uit/ui/components/shadcn/spinner';

import '@blocknote/shadcn/style.css';
import { useAlert } from '@notopia-uit/ui/hooks/use-alert';
import { useIsDocModified } from '@notopia-uit/ui/hooks/use-is-doc-modified';
import { authClient } from '@notopia-uit/ui/lib/auth-client';
import { CloudCheck, CloudUpload, RefreshCw, Wifi, WifiOff } from 'lucide-react';
import { useMemo } from 'react';

import { getDeterministicColor } from './../lib/utils/color';
import { ErrorAlert } from './error-alert';
import { Icons } from './icons';
import { RevisionModal } from './revision-modal';
import { Avatar, AvatarImage, AvatarFallback } from './shadcn/avatar';
import { Badge } from './shadcn/badge';
import { Button } from './shadcn/button';
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from './shadcn/tooltip';
import { SuccessAlert } from './success-alert';

function EditorStatus() {
  const connection = useHocuspocusConnectionStatus();
  const sync = useHocuspocusSyncStatus();
  const users = useHocuspocusAwareness();
  return (
    <TooltipProvider>
      <div className="bg-background/95 supports-backdrop-filter:bg-background/60 sticky top-0 z-50 flex items-center gap-3 border-b px-4 py-2 backdrop-blur-sm">
        <div className="flex items-center gap-2">
          {connection === 'connected' ? (
            <Badge
              variant="secondary"
              className="gap-1 border-emerald-200 bg-emerald-50 px-2 text-emerald-700"
            >
              <Wifi className="size-3.5" />
              Connected
            </Badge>
          ) : connection === 'connecting' ? (
            <Badge variant="outline" className="text-muted-foreground animate-pulse gap-1 px-2">
              <RefreshCw className="size-3.5 animate-spin" />
              Connecting
            </Badge>
          ) : (
            <Badge variant="destructive" className="gap-1 px-2">
              <WifiOff className="size-3.5" />
              Offline
            </Badge>
          )}
        </div>

        <Tooltip>
          <TooltipTrigger asChild>
            <div className="text-muted-foreground hover:text-foreground flex items-center transition-colors">
              {sync === 'syncing' ? (
                <CloudUpload className="size-5 animate-bounce" />
              ) : (
                <CloudCheck className="size-5 text-emerald-500" />
              )}
            </div>
          </TooltipTrigger>
          <TooltipContent>
            {sync === 'syncing' ? 'Saving changes...' : 'All changes saved to server'}
          </TooltipContent>
        </Tooltip>

        <div className="ml-auto flex items-center gap-2">
          <div className="mr-2 flex -space-x-2">
            {users.slice(0, 3).map((user) => (
              <Tooltip key={user.clientId}>
                <TooltipTrigger asChild>
                  <Avatar className="border-background ring-border size-8 border-2 ring-1">
                    <AvatarImage src={user.user.avatar} />
                    <AvatarFallback
                      style={{ backgroundColor: user.user.color }}
                      className="text-[10px] text-white"
                    >
                      {user.user.name.substring(0, 2).toUpperCase() ?? '??'}
                    </AvatarFallback>
                  </Avatar>
                </TooltipTrigger>
                <TooltipContent>{user.user.name}</TooltipContent>
              </Tooltip>
            ))}
          </div>

          {users.length > 3 && (
            <span className="text-muted-foreground text-xs">+{users.length - 3} more</span>
          )}
        </div>
      </div>
    </TooltipProvider>
  );
}

//TODO: set local state for avatar when override from session data, and update awareness state on session change (e.g. login/logout or user update)
//TODO: add dialog
export default function Editor({ noteId }: { noteId: string }) {
  const { data: sessionData } = authClient.useSession();
  const mySchema = useMemo(() => createBlockNoteSchema(), []);
  const provider = useHocuspocusProvider();

  const { isModified, setModified } = useIsDocModified(
    provider.document,
    provider.awareness?.clientID.toString() ?? 'anonymous'
  );

  const { alert, showAlert } = useAlert();

  const { mutate: commitDocument, isPending: isCommitingDocument } = useCommitDocumentMutation({
    onSuccess: (responses) => {
      setModified(false);
      showAlert(
        'success',
        'Document committed successfully!',
        `Your note ${responses.id} changes were committed.`
      );
    },
    onError: (error) => {
      showAlert(
        'error',
        'Failed to commit document',
        `An error occurred while committing your note changes. ${
          error instanceof Error ? error.message : 'Please try again.'
        }`
      );
    },
  });
  const editor = useCreateBlockNote({
    schema: mySchema,
    collaboration: {
      provider: {
        awareness: provider.awareness ? provider.awareness : undefined,
      },
      fragment: provider.document.getXmlFragment('document-store'),
      user: {
        name: sessionData?.user?.name ?? 'Anonymous',
        color: getDeterministicColor(sessionData?.user?.id ?? 'anonymous'),
        avatar: sessionData?.user?.image ?? 'https://placehold.net/default.svg',
      },
    },
  });

  return (
    <div className="relative min-h-screen">
      <EditorStatus />
      <div className="bg-background/95 supports-backdrop-filter:bg-background/60 sticky top-14 z-40 flex gap-2 border-b px-4 py-2 backdrop-blur-sm">
        <RevisionModal noteId={noteId} currentEditor={editor as any} />
      </div>
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
              onClick={() => {
                commitDocument({
                  path: {
                    documentId: noteId,
                  },
                });
              }}
            >
              <Icons.Save />
            </Button>
          )}
        </div>
      )}
    </div>
  );
}
