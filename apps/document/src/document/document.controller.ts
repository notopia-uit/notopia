import { Injectable } from '@nestjs/common';
import {
  DocumentApi,
  GetDocumentAttachmentUploadUrl200Response,
} from '@notopia-uit/api-document-nestjs-server';

import { DocumentService } from './document.service';

@Injectable()
export class DocumentController extends DocumentApi {
  constructor(private readonly documentService: DocumentService) {
    super();
  }

  async getDocumentAttachmentUploadUrl(
    documentId: string,
    _request: Request
  ): Promise<GetDocumentAttachmentUploadUrl200Response> {
    const result =
      await this.documentService.getAttachmentUploadUrl(documentId);
    return { url: result.url };
  }

  async importDocuments(
    _requestBody: Array<object>,
    _request: Request
  ): Promise<void> {
    return;
  }

  async wsDocument(_documentId: string, _request: Request): Promise<void> {
    return;
  }
}
