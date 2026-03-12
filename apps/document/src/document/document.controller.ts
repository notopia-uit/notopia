import { Injectable, UseGuards } from '@nestjs/common';
import {
  DocumentApi,
  GetDocumentAttachmentUploadUrl200Response,
} from '@notopia-uit/api-document-nestjs-server';
import { Traceable } from 'nestjs-otel';

import { HttpUserGuard } from '../common/user.guard';
import { DocumentService } from './document.service';

@Injectable()
@UseGuards(HttpUserGuard)
@Traceable()
export class DocumentController extends DocumentApi {
  constructor(private readonly documentService: DocumentService) {
    super();
  }

  async getDocumentAttachmentUploadUrl(
    documentId: string,
    req: Request
  ): Promise<GetDocumentAttachmentUploadUrl200Response> {
    const result =
      await this.documentService.getAttachmentUploadUrl(documentId);
    return { url: result.url };
  }

  async importDocuments(reqBody: Array<object>, req: Request): Promise<void> {
    return;
  }
}
