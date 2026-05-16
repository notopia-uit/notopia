import { NotFoundException } from '@nestjs/common';

export class DocumentNotFoundException extends NotFoundException {
  override name = DocumentNotFoundException.name;

  readonly documentId: string;
  constructor(documentId: string, cause?: unknown) {
    super(`Document ${documentId} not found`, { cause });
    this.documentId = documentId;
  }
}
