import { Module } from '@nestjs/common';
import { createSchema } from '@notopia-uit/lib-server/block-note';

import { BLOCKNOTE_SCHEMA } from './token';

@Module({
  imports: [],
  providers: [
    {
      provide: BLOCKNOTE_SCHEMA,
      useFactory: createSchema,
    },
  ],
  exports: [BLOCKNOTE_SCHEMA],
})
export class BlockNoteModule {}
