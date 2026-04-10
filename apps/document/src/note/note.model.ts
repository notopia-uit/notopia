export type NoteModel = {
  id: string;
  name: string;
  icon?: string;
  folderId: string;
  tags: string[];
  updatedAt?: Date;
  trashed?: TrashedModel;
};

export type TrashedByModel = 'purpose' | 'parent';

export type TrashedModel = {
  by: TrashedByModel;
  at: Date;
};
