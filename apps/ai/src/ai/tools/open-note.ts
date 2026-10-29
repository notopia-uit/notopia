import { AiError } from '#/common/ai-error';

import type { CustomToolDeps } from './custom-tools';

export async function authorizeEditingTarget(
  deps: Pick<
    CustomToolDeps,
    'userId' | 'workspaceId' | 'authorization' | 'noteClient' | 'abortSignal'
  >,
  noteId: string,
  currentTarget?: string
): Promise<string> {
  deps.abortSignal.throwIfAborted();
  if (currentTarget && currentTarget !== noteId) {
    throw new AiError(409, 'invalidAssistantTarget', 'An assistant task can edit only one note.');
  }
  const permissions = await deps.authorization.permissions(deps.userId, deps.workspaceId);
  if (!permissions.canRead || !permissions.canWrite) {
    throw new AiError(403, 'forbidden', 'Editing tools require workspace write permission.');
  }
  const note = await deps.noteClient.getNote({ userId: deps.userId, noteId });
  if (!note) throw new AiError(404, 'noteNotFound', 'The note is missing or trashed.');
  const workspace = await deps.noteClient.getNoteWorkspace(deps.userId, noteId);
  if (workspace?.id !== deps.workspaceId) {
    throw new AiError(403, 'forbidden', 'Note is outside this workspace.');
  }
  deps.abortSignal.throwIfAborted();
  return noteId;
}
