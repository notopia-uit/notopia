import { NotFoundException } from '@nestjs/common';

export class RevisionNotFoundException extends NotFoundException {
  override name = RevisionNotFoundException.name;

  readonly revisionId: string;
  constructor(revisionId: string, cause?: unknown) {
    super(`Revision ${revisionId} not found`, { cause });
    this.revisionId = revisionId;
  }
}
