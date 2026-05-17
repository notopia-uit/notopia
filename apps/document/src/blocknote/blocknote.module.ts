import { Module } from '@nestjs/common';
import { createServerBlockNoteSchema } from '@notopia-uit/lib/server';

import { BlocknoteService } from './blocknote.service';
import { BLOCKNOTE_SCHEMA } from './token';

@Module({
  imports: [],
  providers: [
    {
      provide: BLOCKNOTE_SCHEMA,
      useFactory: createServerBlockNoteSchema,
    },
    BlocknoteService,
  ],
  exports: [BLOCKNOTE_SCHEMA, BlocknoteService],
})
export class BlockNoteModule {}
