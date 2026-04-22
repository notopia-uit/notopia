export type UserNotePermissions = 'read' | 'write' | 'delete';

export type WorkspaceItemPermission = {
  canRead: boolean;
  canWrite: boolean;
  canDelete: boolean;
};
