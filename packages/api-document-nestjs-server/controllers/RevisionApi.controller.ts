import { Body, Controller, DefaultValuePipe, Delete, Get, Post, Param, ParseIntPipe, ParseFloatPipe, Query, Req } from '@nestjs/common';
import { Observable } from 'rxjs';
import { Cookies, Headers } from '../decorators';
import { RevisionApi } from '../api';
import { GetRevisions200Response, RenameRevisionRequest, Revision,  } from '../models';

@Controller()
export class RevisionApiController {
  constructor(private readonly revisionApi: RevisionApi) {}

  @Delete('/document/revisions/:revisionId')
  deleteRevision(@Param('revisionId') revisionId: string, @Req() request: Request): void | Promise<void> | Observable<void> {
    return this.revisionApi.deleteRevision(revisionId, request);
  }

  @Get('/document/revisions/:revisionId')
  getRevision(@Param('revisionId') revisionId: string, @Req() request: Request): Revision | Promise<Revision> | Observable<Revision> {
    return this.revisionApi.getRevision(revisionId, request);
  }

  @Get('/document/revisions')
  getRevisions(@Query('documentId') documentId: string, @Query('page', new DefaultValuePipe(1)) page: number | undefined, @Query('limit', new DefaultValuePipe(20)) limit: number | undefined, @Req() request: Request): GetRevisions200Response | Promise<GetRevisions200Response> | Observable<GetRevisions200Response> {
    return this.revisionApi.getRevisions(documentId, page, limit, request);
  }

  @Post('/document/revisions/:revisionId/rename')
  renameRevision(@Param('revisionId') revisionId: string, @Body() renameRevisionRequest: RenameRevisionRequest, @Req() request: Request): void | Promise<void> | Observable<void> {
    return this.revisionApi.renameRevision(revisionId, renameRevisionRequest, request);
  }

} 