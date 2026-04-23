import { Module } from '@nestjs/common';
import { createServerBlockNoteSchema } from '@notopia-uit/lib/server';

import { NoteModule } from '#/note/note.module';

export const BLOCKNOTE_SCHEMA = Symbol('BLOCKNOTE_SCHEMA');

@Module({
  imports: [NoteModule],
  providers: [
    {
      provide: BLOCKNOTE_SCHEMA,
      useFactory: createServerBlockNoteSchema,
    },
  ],
  exports: [BLOCKNOTE_SCHEMA],
})
export class BlockNoteModule {}
