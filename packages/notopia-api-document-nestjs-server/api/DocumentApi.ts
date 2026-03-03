import { Injectable } from '@nestjs/common';
import { Observable } from 'rxjs';
import { GetDocumentAttachmentUrl200Response,  } from '../models';


@Injectable()
export abstract class DocumentApi {

  abstract getDocumentAttachmentUrl(documentId: string,  request: Request): GetDocumentAttachmentUrl200Response | Promise<GetDocumentAttachmentUrl200Response> | Observable<GetDocumentAttachmentUrl200Response>;


  abstract importDocuments(requestBody: Array<object>,  request: Request): void | Promise<void> | Observable<void>;


  abstract wsDocument(documentId: string,  request: Request): void | Promise<void> | Observable<void>;

} 