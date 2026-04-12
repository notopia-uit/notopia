import { NotFoundException } from '@nestjs/common';

export class WorkspaceNoteNotFoundException extends NotFoundException {
  constructor(noteId: string) {
    super(`No workspace contains a note with id ${noteId}`);
  }
}
