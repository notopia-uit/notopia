import { Body, Controller, Get, Param, Query, Req } from '@nestjs/common';
import { Observable } from 'rxjs';
import { EditApi } from '../api';
import {  } from '../models';

@Controller()
export class EditApiController {
  constructor(private readonly editApi: EditApi) {}

  @Get('/ws/edits/:documentId')
  wsEditsDocument(@Param('documentId') documentId: string, @Req() request: Request): void | Promise<void> | Observable<void> {
    return this.editApi.wsEditsDocument(documentId, request);
  }

} 