import { User } from '../common/user';
import { DocumentService } from './document.service';
import { UnauthorizedException  } from '@nestjs/common';
import { DocumentApi as DocumentApiDefinition } from '@notopia-uit/api-document-nestjs-server/api';
import { Traceable } from 'nestjs-otel';

@Traceable()
export class DocumentApi extends DocumentApiDefinition {
  constructor(private readonly documentService: DocumentService) {
    super();
  }

  async commitDocument(documentId: string, _: Request) {
    await this.documentService.commitDocument(documentId);
  }

  async getDocumentAttachmentUploadUrl(documentId: string, req: Request) {
    const user = (req as any).user as User | undefined;
    if (!user) {
      throw new UnauthorizedException('User not authenticated');
    }
    return await this.documentService.getAttachmentUploadUrl(documentId, user);
  }
}
