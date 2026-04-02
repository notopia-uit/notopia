import { Injectable } from '@nestjs/common';
import { Observable } from 'rxjs';
import { GetRevisions200Response, RenameRevisionRequest, Revision,  } from '../models';


@Injectable()
export abstract class RevisionApi {

  abstract deleteRevision(revisionId: string,  request: Request): void | Promise<void> | Observable<void>;


  abstract getRevision(revisionId: string,  request: Request): Revision | Promise<Revision> | Observable<Revision>;


  abstract getRevisions(documentId: string, page: number | undefined, limit: number | undefined,  request: Request): GetRevisions200Response | Promise<GetRevisions200Response> | Observable<GetRevisions200Response>;


  abstract renameRevision(revisionId: string, renameRevisionRequest: RenameRevisionRequest,  request: Request): void | Promise<void> | Observable<void>;

} 