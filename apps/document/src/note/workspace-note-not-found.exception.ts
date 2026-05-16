import { NotFoundException } from '@nestjs/common';

export class WorkspaceNoteNotFoundException extends NotFoundException {
  override name = WorkspaceNoteNotFoundException.name;

  readonly noteId: string;

  constructor(noteId: string, cause?: unknown) {
    super(`No workspace contains a note with id ${noteId}`, { cause });
    this.noteId = noteId;
  }
}
