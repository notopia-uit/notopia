import { Injectable } from '@nestjs/common';
import { Observable } from 'rxjs';
import {  } from '../models';


@Injectable()
export abstract class DocumentApi {

  abstract wsDocument(documentId: string,  request: Request): void | Promise<void> | Observable<void>;

} 