import { CollaborationUser as BlockNoteCollaborationUser } from '@blocknote/core/extensions';
import { CollabUser as HocuspocusCollabUser } from '@hocuspocus/provider-react';

export type CollabUser = {
  avatar: string;
} & BlockNoteCollaborationUser &
  HocuspocusCollabUser;
