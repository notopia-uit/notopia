import { Body, Controller, Get, Param, Query, Req } from '@nestjs/common';
import { Observable } from 'rxjs';
import { DocumentApi } from '../api';
import {  } from '../models';

@Controller()
export class DocumentApiController {
  constructor(private readonly documentApi: DocumentApi) {}

  @Get('/ws/documents/:documentId')
  wsDocument(@Param('documentId') documentId: string, @Req() request: Request): void | Promise<void> | Observable<void> {
    return this.documentApi.wsDocument(documentId, request);
  }

} 