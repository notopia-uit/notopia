import { Injectable } from '@nestjs/common';
import { Observable } from 'rxjs';
import { GetDocumentAttachmentUploadUrl200Response,  } from '../models';


@Injectable()
export abstract class DocumentApi {

  abstract commitDocument(documentId: string,  request: Request): void | Promise<void> | Observable<void>;


  abstract getDocumentAttachmentUploadUrl(documentId: string,  request: Request): GetDocumentAttachmentUploadUrl200Response | Promise<GetDocumentAttachmentUploadUrl200Response> | Observable<GetDocumentAttachmentUploadUrl200Response>;

} 