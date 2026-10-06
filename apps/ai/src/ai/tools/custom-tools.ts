import { tool } from 'ai';
import { z } from 'zod';

import type { AuthorizationService } from '#/authorization/authorization.service';
import type { NoteService } from '#/note/note.service';
import type { SearchService } from '#/search/search.service';

export type CustomToolDeps = {
  userId: string;
  noteClient: NoteService;
  search: SearchService;
  authorization: AuthorizationService;
};

export function createCustomTools(deps: CustomToolDeps) {
  return {
    searchNotes: tool({
      description:
        "Full-text search over the user's notes by title, content, or tags. Returns matching notes with their IDs. Call getNote for details on a specific match.",
      inputSchema: z.object({
        query: z.string().describe('The search query, e.g. a topic, title, or keyword.'),
        workspaceId: z
          .string()
          .optional()
          .describe('Restrict results to this workspace ID, if known.'),
        limit: z.number().int().min(1).max(20).default(5),
      }),
      execute: async ({ query, workspaceId, limit }) =>
        deps.search.searchNotes({ query, workspaceId, limit }),
    }),
    getNote: tool({
      description:
        'Fetch a single note by ID: its title, tags, and folder. Returns null when the note does not exist or is trashed.',
      inputSchema: z.object({
        noteId: z.string().describe('The note ID, e.g. from a searchNotes result.'),
      }),
      execute: async ({ noteId }) => deps.noteClient.getNote({ userId: deps.userId, noteId }),
    }),
    listWorkspaces: tool({
      description:
        'List the workspaces the user belongs to, with their role in each. Use this to discover workspace IDs for searchNotes.',
      inputSchema: z.object({}),
      execute: async () => deps.authorization.getUserWorkspaces({ userId: deps.userId }),
    }),
  };
}

export type CustomTools = ReturnType<typeof createCustomTools>;
