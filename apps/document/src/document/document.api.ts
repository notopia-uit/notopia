import { UnauthorizedException } from '@nestjs/common';
import { DocumentApi as DocumentApiDefinition } from '@notopia-uit/api-document-nestjs-server/api';
import { Traceable } from 'nestjs-otel';

import { User } from '../common/user';
import { DocumentService } from './document.service';

@Traceable()
export class DocumentApi extends DocumentApiDefinition {
  constructor(private readonly documentService: DocumentService) {
    super();
  }

  async commitDocument(documentId: string, req: Request) {
    const user = (req as unknown as Record<string, unknown>).user as User | undefined;
    if (!user) {
      throw new UnauthorizedException('User not authenticated');
    }
    await this.documentService.commitDocument({ documentId, userId: user.id });
  }

  async getDocumentAttachmentUploadUrl(documentId: string, req: Request) {
    const user = (req as unknown as Record<string, unknown>).user as User | undefined;
    if (!user) {
      throw new UnauthorizedException('User not authenticated');
    }
    return await this.documentService.getAttachmentUploadUrl(documentId, user.id);
  }
}
