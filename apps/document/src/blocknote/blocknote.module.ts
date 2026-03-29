import {
  BLOCKNOTE_SCHEMA,
  BlockNoteSchemaProvider,
} from './bn-schema.provider';
import { Module } from '@nestjs/common';

@Module({
  providers: [
    {
      provide: BLOCKNOTE_SCHEMA,
      useFactory: BlockNoteSchemaProvider,
    },
  ],
  exports: [BLOCKNOTE_SCHEMA],
})
export class BlockNoteModule {}
