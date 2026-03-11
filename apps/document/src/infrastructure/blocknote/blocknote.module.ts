import { ServerBlockNoteEditor } from '@blocknote/server-util';
import { Module } from '@nestjs/common';

@Module({
  providers: [
    {
      provide: ServerBlockNoteEditor,
      useFactory: () => {
        return ServerBlockNoteEditor.create();
      },
    },
  ],
  exports: [ServerBlockNoteEditor],
})
export class BlockNoteModule {}
