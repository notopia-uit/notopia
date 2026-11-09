import { Logger } from '@nestjs/common';
import { tool } from 'ai';
import { z } from 'zod';

import type { AuthorizationService } from '#/authorization/authorization.service';
import { AiError, publicError } from '#/common/ai-error';
import type { NoteService } from '#/note/note.service';
import type { SearchService } from '#/search/search.service';

export type CustomToolDeps = {
  userId: string;
  workspaceId: string;
  noteClient: NoteService;
  search: SearchService;
  authorization: AuthorizationService;
  abortSignal: AbortSignal;
};

export function createCustomTools(deps: CustomToolDeps) {
  const logger = new Logger('AiTools');
  const checkAccess = async () => {
    deps.abortSignal.throwIfAborted();
    if (!(await deps.authorization.canRead(deps.userId, deps.workspaceId))) {
      throw new AiError(403, 'forbidden', 'Workspace access denied.');
    }
    deps.abortSignal.throwIfAborted();
  };
  const execute =
    <T>(fn: () => Promise<T>) =>
    async () => {
      try {
        await checkAccess();
        return await fn();
      } catch (error) {
        logger.warn({ err: error }, 'AI tool failed');
        throw new Error(String(publicError(error).body.message));
      }
    };
  const getNote = async (noteId: string) => {
    const note = await deps.noteClient.getNote({ userId: deps.userId, noteId });
    if (!note) return null;
    const workspace = await deps.noteClient.getNoteWorkspace(deps.userId, noteId);
    if (workspace?.id !== deps.workspaceId)
      throw new AiError(403, 'forbidden', 'Note is outside this workspace.');
    return note;
  };
  return {
    searchNotes: tool({
      description:
        'Search notes in the current workspace. Returns committed content excerpts, which can lag behind live edits.',
      inputSchema: z.object({
        query: z.string().min(1).max(1000),
        limit: z.number().int().min(1).max(20).default(5),
      }),
      execute: ({ query, limit }) =>
        execute(async () => {
          const hits = await deps.search.searchNotes({
            query,
            limit,
            workspaceId: deps.workspaceId,
          });
          const checked = await Promise.all(
            hits.map(async (hit) => ((await getNote(hit.id)) ? hit : null))
          );
          return checked.filter((hit) => hit !== null);
        })(),
    }),
    getNote: tool({
      description:
        'Read note metadata (title, tags, folder), not its content. Returns null for a missing or trashed note.',
      inputSchema: z.object({ noteId: z.uuid() }),
      execute: ({ noteId }) => execute(() => getNote(noteId))(),
    }),
    getNoteWorkspace: tool({
      description: 'Read workspace details for a note in the current workspace.',
      inputSchema: z.object({ noteId: z.uuid() }),
      execute: ({ noteId }) =>
        execute(async () => {
          if (!(await getNote(noteId))) return null;
          return deps.noteClient.getNoteWorkspace(deps.userId, noteId);
        })(),
    }),
    readIndexedNoteContent: tool({
      description:
        'Read the last indexed committed note content. This is not the live editor state and may be stale or truncated.',
      inputSchema: z.object({ noteId: z.uuid() }),
      execute: ({ noteId }) =>
        execute(async () => {
          if (!(await getNote(noteId))) return null;
          return deps.search.readIndexedNoteContent(noteId, deps.workspaceId);
        })(),
    }),
    listWorkspaces: tool({
      description:
        'List accessible workspace IDs and roles. Other tools remain scoped to the current workspace.',
      inputSchema: z.object({}),
      execute: execute(() => deps.authorization.getUserWorkspaces({ userId: deps.userId })),
    }),
  };
}

export type CustomTools = ReturnType<typeof createCustomTools>;
