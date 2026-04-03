import { AppConfig } from '../config/config';
import { APP_CONFIG } from '../config/config.factory';
import { NoteModule } from '../note/note.module';
import { NoteService } from '../note/note.service';
import { Module } from '@nestjs/common';
import { ConfigService } from '@nestjs/config';
import { createBlockNoteSchema } from '@notopia-uit/block-note';

export const BLOCKNOTE_SCHEMA = Symbol('BLOCKNOTE_SCHEMA');

@Module({
  imports: [NoteModule],
  providers: [
    {
      provide: BLOCKNOTE_SCHEMA,
      useFactory: (noteService: NoteService, configService: ConfigService) => {
        const appCfg = configService.get<AppConfig>(APP_CONFIG)!;
        const getNoteName = async (noteId: string) => {
          const noteName = await noteService.getNoteName(noteId);
          return noteName;
        };

        return createBlockNoteSchema({ baseUrl: appCfg.apiUrl, getNoteName });
      },
      inject: [NoteService, ConfigService],
    },
  ],
  exports: [BLOCKNOTE_SCHEMA],
})
export class BlockNoteModule {}
