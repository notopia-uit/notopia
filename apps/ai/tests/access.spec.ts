import { describe, expect, it, vi } from 'vitest';

import { AiService } from '../src/ai/ai.service';
import type { BlockNoteChatBody } from '../src/ai/chat-request.dto';
import { createCustomTools } from '../src/ai/tools/custom-tools';
import { authorizeEditingTarget } from '../src/ai/tools/open-note';

const workspaceId = '00000000-0000-4000-8000-000000000001';
const other = '00000000-0000-4000-8000-000000000002';
const noteId = '00000000-0000-4000-8000-000000000003';
const makeDeps = () => ({
  userId: 'user',
  workspaceId,
  abortSignal: new AbortController().signal,
  authorization: {
    canRead: vi.fn().mockResolvedValue(true),
    permissions: vi.fn().mockResolvedValue({ canRead: true, canWrite: true }),
    getUserWorkspaces: vi.fn().mockResolvedValue([{ workspaceId }]),
  },
  noteClient: {
    getNoteWorkspace: vi.fn().mockResolvedValue({ id: workspaceId }),
    getNote: vi.fn().mockResolvedValue({ id: noteId }),
  },
  search: {
    searchNotes: vi.fn().mockResolvedValue([{ id: noteId }]),
    readIndexedNoteContent: vi.fn().mockResolvedValue({ content: 'hello' }),
  },
});
const body = (extra = {}) => ({ messages: [], ...extra }) as BlockNoteChatBody;
const service = (deps: ReturnType<typeof makeDeps>) =>
  new AiService(
    {} as any,
    {} as any,
    deps.noteClient as any,
    deps.search as any,
    deps.authorization as any,
    {} as any,
    {} as any
  );
const options = { toolCallId: 'call', messages: [] };

describe('AI workspace authorization', () => {
  it('authorizes a live editing target in the current workspace', async () => {
    expect(await authorizeEditingTarget(makeDeps() as any, noteId)).toBe(noteId);
  });
  it('rejects opening for viewers', async () => {
    const deps = makeDeps();
    deps.authorization.permissions.mockResolvedValue({ canRead: true, canWrite: false });
    await expect(authorizeEditingTarget(deps as any, noteId)).rejects.toMatchObject({
      status: 403,
    });
  });
  it('rejects missing or trashed editing targets', async () => {
    const deps = makeDeps();
    deps.noteClient.getNote.mockResolvedValue(null);
    await expect(authorizeEditingTarget(deps as any, noteId)).rejects.toMatchObject({
      status: 404,
    });
  });
  it('rejects cross-workspace editing targets', async () => {
    const deps = makeDeps();
    deps.noteClient.getNoteWorkspace.mockResolvedValue({ id: other });
    await expect(authorizeEditingTarget(deps as any, noteId)).rejects.toMatchObject({
      status: 403,
    });
  });
  it('does not switch an existing task target', async () => {
    await expect(authorizeEditingTarget(makeDeps() as any, noteId, other)).rejects.toMatchObject({
      status: 409,
    });
  });
  it('uses the sole workspace for existing clients', async () => {
    const deps = makeDeps();
    expect(await service(deps).resolveWorkspace(body(), 'user')).toBe(workspaceId);
  });
  it('does not guess a billing workspace', async () => {
    const deps = makeDeps();
    deps.authorization.getUserWorkspaces.mockResolvedValue([
      { workspaceId },
      { workspaceId: other },
    ]);
    await expect(service(deps).resolveWorkspace(body(), 'user')).rejects.toMatchObject({
      status: 400,
    });
  });
  it('rejects conflicting workspace and note context', async () => {
    const deps = makeDeps();
    await expect(
      service(deps).resolveWorkspace(body({ workspaceId: other, noteId }), 'user')
    ).rejects.toMatchObject({ status: 403 });
    await expect(
      service(deps).resolveWorkspace(body({ workspaceId }), 'user', other)
    ).rejects.toMatchObject({ status: 400 });
  });
  it('rejects viewers supplying editing tools', async () => {
    const deps = makeDeps();
    deps.authorization.permissions.mockResolvedValue({ canRead: true, canWrite: false });
    await expect(
      service(deps).resolveWorkspace(
        body({ workspaceId, toolDefinitions: { applyDocumentOperations: {} } }),
        'user'
      )
    ).rejects.toMatchObject({ status: 403 });
  });
  it('rechecks permissions before executing tools', async () => {
    const deps = makeDeps();
    deps.authorization.canRead.mockResolvedValue(false);
    const tools = createCustomTools(deps as any);
    await expect(tools.searchNotes.execute!({ query: 'test', limit: 5 }, options)).rejects.toThrow(
      'access denied'
    );
    expect(deps.search.searchNotes).not.toHaveBeenCalled();
  });
  it('never reads indexed content from a note in another workspace', async () => {
    const deps = makeDeps();
    deps.noteClient.getNoteWorkspace.mockResolvedValue({ id: other });
    const tools = createCustomTools(deps as any);
    await expect(tools.readIndexedNoteContent.execute!({ noteId }, options)).rejects.toThrow(
      'outside this workspace'
    );
    expect(deps.search.readIndexedNoteContent).not.toHaveBeenCalled();
  });
  it('filters stale search hits using the live note service', async () => {
    const deps = makeDeps();
    deps.noteClient.getNote.mockResolvedValue(null);
    const tools = createCustomTools(deps as any);
    expect(await tools.searchNotes.execute!({ query: 'test', limit: 5 }, options)).toEqual([]);
    expect(deps.search.searchNotes).toHaveBeenCalledWith({ query: 'test', limit: 5, workspaceId });
  });
});
