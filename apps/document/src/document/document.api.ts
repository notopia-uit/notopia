import { Injectable, UseGuards } from '@nestjs/common';
import { DocumentApi as _DocumentApi } from '@notopia-uit/api-document-nestjs-server';
import { Traceable } from 'nestjs-otel';

import { User } from '../common/user';
import { HttpUserGuard } from '../common/user.guard';
import { DocumentService } from './document.service';

@Injectable()
@UseGuards(HttpUserGuard)
@Traceable()
export class DocumentApi extends _DocumentApi {
  constructor(private readonly documentService: DocumentService) {
    super();
  }

  async commitDocument(documentId: string, _: Request) {
    await this.documentService.commitDocument(documentId);
  }

  async getDocumentAttachmentUploadUrl(documentId: string, req: Request) {
    const user = (req as any).user as User;
    return await this.documentService.getAttachmentUploadUrl(documentId, user);
  }

  async importDocuments(_: Array<object>, __: Request) {
    return;
  }
}
