import { Body, Controller, DefaultValuePipe, Get, Post, Param, ParseIntPipe, ParseFloatPipe, Query, Req } from '@nestjs/common';
import { Observable } from 'rxjs';
import { Cookies, Headers } from '../decorators';
import { DocumentApi } from '../api';
import { CommitDocument201Response, GetDocumentAttachmentUploadUrl200Response,  } from '../models';

@Controller()
export class DocumentApiController {
  constructor(private readonly documentApi: DocumentApi) {}

  @Post('/document/documents/:documentId/commit')
  commitDocument(@Param('documentId') documentId: string, @Req() request: Request): CommitDocument201Response | Promise<CommitDocument201Response> | Observable<CommitDocument201Response> {
    return this.documentApi.commitDocument(documentId, request);
  }

  @Get('/document/documents/:documentId/attachment-url')
  getDocumentAttachmentUploadUrl(@Param('documentId') documentId: string, @Query('filename') filename: string, @Req() request: Request): GetDocumentAttachmentUploadUrl200Response | Promise<GetDocumentAttachmentUploadUrl200Response> | Observable<GetDocumentAttachmentUploadUrl200Response> {
    return this.documentApi.getDocumentAttachmentUploadUrl(documentId, filename, request);
  }

} 