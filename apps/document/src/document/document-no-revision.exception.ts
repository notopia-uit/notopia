import { NotFoundException } from '@nestjs/common';

export class DocumentNoRevisionException extends NotFoundException {
  override name = DocumentNoRevisionException.name;

  readonly documentId: string;
  constructor(documentId: string, cause?: unknown) {
    super(`Document ${documentId} has no changes to commit and no revision yet`, { cause });
    this.documentId = documentId;
  }
}
