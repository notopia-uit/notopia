import { ShareNoteSearch } from '@notopia-uit/api-gen';

export type NoteSearch = Partial<ShareNoteSearch> &
  Required<Pick<ShareNoteSearch, 'id'>>;
