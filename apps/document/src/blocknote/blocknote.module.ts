import { Module } from '@nestjs/common';
import { createSchema } from '@notopia-uit/lib-server/block-note';

import { BlocknoteService } from './blocknote.service';
import { BLOCKNOTE_SCHEMA } from './token';

@Module({
  imports: [],
  providers: [
    {
      provide: BLOCKNOTE_SCHEMA,
      useFactory: createSchema,
    },
    BlocknoteService,
  ],
  exports: [BLOCKNOTE_SCHEMA, BlocknoteService],
})
export class BlockNoteModule {}
