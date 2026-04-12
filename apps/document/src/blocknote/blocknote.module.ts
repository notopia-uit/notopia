import { MySchema } from '@blocknote/core';
import { Module } from '@nestjs/common';
import { ConfigService } from '@nestjs/config';
import { createServerBlockNoteSchema } from '@notopia-uit/lib/server';

import { AppConfig } from '#/config/config';
import { APP_CONFIG } from '#/config/config.factory';
import { NoteModule } from '#/note/note.module';

export const BLOCKNOTE_SCHEMA = Symbol('BLOCKNOTE_SCHEMA');

@Module({
  imports: [NoteModule],
  providers: [
    {
      provide: BLOCKNOTE_SCHEMA,
      useFactory: (configService: ConfigService): MySchema => {
        const appCfg = configService.get<AppConfig>(APP_CONFIG);
        if (!appCfg) {
          throw new Error('APP_CONFIG not found');
        }
        return createServerBlockNoteSchema();
      },
      inject: [ConfigService],
    },
  ],
  exports: [BLOCKNOTE_SCHEMA],
})
export class BlockNoteModule {}
