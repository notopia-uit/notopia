import { NotFoundException } from '@nestjs/common';

export class NoteNotFoundException extends NotFoundException {
  override name = NoteNotFoundException.name;

  readonly noteId: string;
  constructor(noteId: string, cause?: unknown) {
    super(`Note with id ${noteId} not found`, { cause });
    this.noteId = noteId;
  }
}
