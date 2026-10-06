import { CollaborationUser as BlockNoteCollaborationUser } from '@blocknote/core/yjs';

export interface CollaborationUser extends BlockNoteCollaborationUser {
  avatar: string;
}
