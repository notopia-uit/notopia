import { NotFoundException } from '@nestjs/common';

export class NoteNotFoundException extends NotFoundException {
  constructor(noteId: string) {
    super(`Note with id ${noteId} not found`);
  }
}
