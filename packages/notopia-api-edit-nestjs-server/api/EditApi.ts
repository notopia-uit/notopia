import { Injectable } from '@nestjs/common';
import { Observable } from 'rxjs';
import {  } from '../models';


@Injectable()
export abstract class EditApi {

  abstract wsEditsDocument(documentId: string,  request: Request): void | Promise<void> | Observable<void>;

} 