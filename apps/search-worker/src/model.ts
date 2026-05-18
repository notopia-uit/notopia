import { NoteSearch as ShareNoteSearch } from '@notopia-uit/api-share-gen';

export type NoteSearch = Partial<ShareNoteSearch> & Required<Pick<ShareNoteSearch, 'id'>>;
