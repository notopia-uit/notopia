import { Injectable } from '@nestjs/common';
import { Observable } from 'rxjs';
import { CommittedRevision, GetDocumentAttachmentUploadUrl200Response,  } from '../models';


@Injectable()
export abstract class DocumentApi {

  abstract commitDocument(documentId: string,  request: Request): CommittedRevision | Promise<CommittedRevision> | Observable<CommittedRevision>;


  abstract getDocumentAttachmentUploadUrl(documentId: string, filename: string,  request: Request): GetDocumentAttachmentUploadUrl200Response | Promise<GetDocumentAttachmentUploadUrl200Response> | Observable<GetDocumentAttachmentUploadUrl200Response>;

} 