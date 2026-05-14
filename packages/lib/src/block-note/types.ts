import { CollaborationUser as BlockNoteCollaborationUser } from '@blocknote/core/extensions';

export interface CollaborationUser extends BlockNoteCollaborationUser {
  avatar: string;
}
