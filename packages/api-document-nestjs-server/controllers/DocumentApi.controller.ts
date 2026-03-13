import { Body, Controller, Get, Post, Param, Query, Req } from '@nestjs/common';
import { Observable } from 'rxjs';
import { DocumentApi } from '../api';
import { GetDocumentAttachmentUploadUrl200Response,  } from '../models';

@Controller()
export class DocumentApiController {
  constructor(private readonly documentApi: DocumentApi) {}

  @Get('/document/documents/:documentId/commit')
  commitDocument(@Param('documentId') documentId: string, @Req() request: Request): void | Promise<void> | Observable<void> {
    return this.documentApi.commitDocument(documentId, request);
  }

  @Get('/document/documents/:documentId/attachment-url')
  getDocumentAttachmentUploadUrl(@Param('documentId') documentId: string, @Req() request: Request): GetDocumentAttachmentUploadUrl200Response | Promise<GetDocumentAttachmentUploadUrl200Response> | Observable<GetDocumentAttachmentUploadUrl200Response> {
    return this.documentApi.getDocumentAttachmentUploadUrl(documentId, request);
  }

  @Post('/document/documents/import')
  importDocuments(@Body() requestBody: Array<object>, @Req() request: Request): void | Promise<void> | Observable<void> {
    return this.documentApi.importDocuments(requestBody, request);
  }

} 