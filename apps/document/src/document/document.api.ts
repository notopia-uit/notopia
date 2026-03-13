import type { Client } from '@connectrpc/connect';
import {
  Inject,
  Injectable,
  UnauthorizedException,
  UseGuards,
} from '@nestjs/common';
import {
  DocumentApi as _DocumentApi,
  GetDocumentAttachmentUploadUrl200Response,
} from '@notopia-uit/api-document-nestjs-server';
import {
  AuthorizationService,
  NotePermission,
} from '@notopia-uit/pb/authorization';
import { Traceable } from 'nestjs-otel';

import { AUTHORIZATION_SERVICE } from '../authorization/authorization.module';
import { User } from '../common/user';
import { HttpUserGuard } from '../common/user.guard';
import { DocumentService } from './document.service';

@Injectable()
@UseGuards(HttpUserGuard)
@Traceable()
export class DocumentApi extends _DocumentApi {
  constructor(
    private readonly documentService: DocumentService,
    @Inject(AUTHORIZATION_SERVICE)
    private readonly authorizationService: Client<typeof AuthorizationService>
  ) {
    super();
  }

  async getDocumentAttachmentUploadUrl(
    documentId: string,
    req: Request
  ): Promise<GetDocumentAttachmentUploadUrl200Response> {
    const user = (req as any).user as User;
    const permissionRes = await this.authorizationService.hasNotePermission({
      noteId: documentId,
      permission: NotePermission.WRITE,
      memberId: user.id,
    });
    if (!permissionRes.hasPermission) {
      throw new UnauthorizedException(
        `User ${user.id} does not have permission to upload attachment to ${documentId}`
      );
    }
    return await this.documentService.getAttachmentUploadUrl(documentId);
  }

  async importDocuments(_: Array<object>, __: Request): Promise<void> {
    return;
  }
}
