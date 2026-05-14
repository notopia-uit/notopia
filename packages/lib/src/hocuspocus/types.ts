import { CollaborationUser as BlockNoteCollaborationUser } from '../block-note/types';
import '@hocuspocus/provider-react';

type PickedBlockNoteCollaborationUser = Pick<
  BlockNoteCollaborationUser,
  'name' | 'color' | 'avatar'
>;

declare module '@hocuspocus/provider-react' {
  interface CollabUser extends PickedBlockNoteCollaborationUser {}
}
