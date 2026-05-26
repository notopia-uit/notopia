'use client';

import { BlockNoteEditor } from '@blocknote/core';
import { SuggestionMenuController, useCreateBlockNote } from '@blocknote/react';
import { BlockNoteView } from '@blocknote/shadcn';
import {
  useHocuspocusProvider,
  useHocuspocusAwareness,
  useHocuspocusSyncStatus,
  useHocuspocusConnectionStatus,
} from '@hocuspocus/provider-react';
import { type CollaborationUser } from '@notopia-uit/lib/block-note';
import {
  createBlockNoteSchema,
  getNoteMenuItems,
  getTagMenuItems,
  searchNotesFromMeilisearch,
  searchTagsFromMeilisearch,
} from '@notopia-uit/ui/block-note';
import { getMenuItemsWithState } from '@notopia-uit/ui/block-note/menu-states';
import { useMeilisearch } from '@notopia-uit/ui/contexts/meilisearch-context';
import { useSearchCache } from '@notopia-uit/ui/hooks/use-search-cache';
import { uploadDocumentAttachment } from '@notopia-uit/ui/lib/actions/upload';
import { CloudCheck, CloudUpload, RefreshCw, Wifi, WifiOff } from 'lucide-react';
import { useTheme } from 'next-themes';
import { forwardRef, useMemo, useCallback, useEffect } from 'react';

import { Avatar, AvatarImage, AvatarFallback } from './shadcn/avatar';
import { Badge } from './shadcn/badge';
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from './shadcn/tooltip';

interface EditorCoreProps {
  sessionUser?: {
    name: string;
    color: string;
    avatar: string;
  };
  noteId: string;
  isViewer?: boolean;
}

function EditorStatusBar() {
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

export const EditorCore = forwardRef<BlockNoteEditor | null, EditorCoreProps>(function EditorCore(
  { sessionUser, noteId, isViewer },
  ref
) {
  const { resolvedTheme } = useTheme();
  const mySchema = useMemo(() => createBlockNoteSchema(), []);
  const provider = useHocuspocusProvider();
  const meilisearchClient = useMeilisearch();

  const uploadFile = useCallback(
    async (file: File): Promise<string> => {
      return uploadDocumentAttachment(noteId, file);
    },
    [noteId]
  );

  const resolveFileUrl = useCallback((url: string): Promise<string> => {
    return Promise.resolve(url);
  }, []);

  const editor = useCreateBlockNote({
    schema: mySchema,
    collaboration: {
      provider: {
        awareness: provider.awareness ? provider.awareness : undefined,
      },
      fragment: provider.document.getXmlFragment('prosemirror'),
      user: (sessionUser ?? {
        name: 'Anonymous',
        color: '#999999',
        avatar: 'https://placehold.net/default.svg',
      }) satisfies CollaborationUser,
    },
    uploadFile,
    resolveFileUrl,
  });

  useEffect(() => {
    if (ref) {
      (ref as React.MutableRefObject<any>).current = editor;
    }
    return () => {
      if (ref) {
        (ref as React.MutableRefObject<any>).current = null;
      }
    };
  }, [editor, ref]);

  const noteSearchFn = useCallback(
    (query: string) => searchNotesFromMeilisearch(meilisearchClient, query),
    [meilisearchClient]
  );

  const tagSearchFn = useCallback(
    (query: string) => searchTagsFromMeilisearch(meilisearchClient, query),
    [meilisearchClient]
  );

  const { search: searchNotesWithCache } = useSearchCache(noteSearchFn);
  const { search: searchTagsWithCache } = useSearchCache(tagSearchFn);

  const handleNoteMenuSearch = useCallback(
    async (query: string) => {
      const result = await searchNotesWithCache(query);
      const menuItems = getNoteMenuItems(editor, query, result.data || []);
      return getMenuItemsWithState({
        items: menuItems,
        isLoading: result.isLoading,
        error: result.error,
        query,
      });
    },
    [searchNotesWithCache, editor]
  );

  const handleTagMenuSearch = useCallback(
    async (query: string) => {
      const result = await searchTagsWithCache(query);
      const menuItems = getTagMenuItems(editor, query, result.data || []);
      return getMenuItemsWithState({
        items: menuItems,
        isLoading: result.isLoading,
        error: result.error,
        query,
      });
    },
    [searchTagsWithCache, editor]
  );

  return (
    <>
      <EditorStatusBar />
      <BlockNoteView editor={editor} theme={resolvedTheme as 'light' | 'dark'} editable={!isViewer}>
        <SuggestionMenuController
          triggerCharacter={'#'}
          getItems={async (query) => {
            return handleTagMenuSearch(query);
          }}
        />

        <SuggestionMenuController
          triggerCharacter={'@'}
          getItems={async (query) => {
            return handleNoteMenuSearch(query);
          }}
        />
      </BlockNoteView>
    </>
  );
});
