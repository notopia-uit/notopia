import { CollaborationUser as BlockNoteCollaborationUser } from '../block-note/types';
import '@hocuspocus/provider-react';

declare module '@hocuspocus/provider-react' {
  interface CollabUser {
    user: BlockNoteCollaborationUser;
  }
}
