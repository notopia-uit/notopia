import { Body, Controller, Get, Post, Param, Query, Req } from '@nestjs/common';
import { Observable } from 'rxjs';
import { DocumentApi } from '../api';
import { GetDocumentAttachmentUrl200Response,  } from '../models';

@Controller()
export class DocumentApiController {
  constructor(private readonly documentApi: DocumentApi) {}

  @Get('/document/documents/:documentId/attachment-url')
  getDocumentAttachmentUrl(@Param('documentId') documentId: string, @Req() request: Request): GetDocumentAttachmentUrl200Response | Promise<GetDocumentAttachmentUrl200Response> | Observable<GetDocumentAttachmentUrl200Response> {
    return this.documentApi.getDocumentAttachmentUrl(documentId, request);
  }

  @Post('/document/documents/import')
  importDocuments(@Body() requestBody: Array<object>, @Req() request: Request): void | Promise<void> | Observable<void> {
    return this.documentApi.importDocuments(requestBody, request);
  }

  @Get('/document/ws/documents/:documentId')
  wsDocument(@Param('documentId') documentId: string, @Req() request: Request): void | Promise<void> | Observable<void> {
    return this.documentApi.wsDocument(documentId, request);
  }

} 