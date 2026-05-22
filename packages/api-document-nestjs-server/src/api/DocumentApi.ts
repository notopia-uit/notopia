import { Injectable } from '@nestjs/common';
import { Observable } from 'rxjs';
import { CommitDocument201Response, GetDocumentAttachmentUploadUrl200Response,  } from '../models';


@Injectable()
export abstract class DocumentApi {

  abstract commitDocument(documentId: string,  request: Request): CommitDocument201Response | Promise<CommitDocument201Response> | Observable<CommitDocument201Response>;


  abstract getDocumentAttachmentUploadUrl(documentId: string, filename: string,  request: Request): GetDocumentAttachmentUploadUrl200Response | Promise<GetDocumentAttachmentUploadUrl200Response> | Observable<GetDocumentAttachmentUploadUrl200Response>;

} 