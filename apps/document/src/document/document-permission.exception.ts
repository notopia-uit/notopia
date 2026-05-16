import { UnauthorizedException } from '@nestjs/common';

export class DocumentPermissionException extends UnauthorizedException {
  override name = DocumentPermissionException.name;

  readonly documentId: string;
  readonly userId: string;
  constructor(documentId: string, userId: string, cause?: unknown) {
    super(`User ${userId} does not have permission to access document ${documentId}`, { cause });
    this.documentId = documentId;
    this.userId = userId;
  }
}
