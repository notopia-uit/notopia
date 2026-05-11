import { Logger, UnauthorizedException } from '@nestjs/common';
import { DocumentApi as DocumentApiDefinition } from '@notopia-uit/api-document-nestjs-server/api';
import { CommitDocument201Response } from '@notopia-uit/api-document-nestjs-server/models';
import { Traceable } from 'nestjs-otel';

import { User } from '../common/user';
import { DocumentService } from './document.service';

@Traceable()
export class DocumentApi extends DocumentApiDefinition {
  private readonly logger = new Logger(DocumentApi.name);

  constructor(private readonly documentService: DocumentService) {
    super();
  }

  async commitDocument(documentId: string, req: Request): Promise<CommitDocument201Response> {
    this.logger.log(`commitDocument: received documentId=${documentId}`);
    const user = (req as unknown as Record<string, unknown>).user as User | undefined;
    if (!user) {
      throw new UnauthorizedException('User not authenticated');
    }
    const revisionId = await this.documentService.commitDocument({ documentId, userId: user.id });
    const response = { id: revisionId } as CommitDocument201Response;
    this.logger.log(`commitDocument: done documentId=${documentId}`);
    this.logger.debug({ response }, 'commitDocument: response');
    return response;
  }

  async getDocumentAttachmentUploadUrl(documentId: string, req: Request) {
    this.logger.log(`getDocumentAttachmentUploadUrl: received documentId=${documentId}`);
    const user = (req as unknown as Record<string, unknown>).user as User | undefined;
    if (!user) {
      throw new UnauthorizedException('User not authenticated');
    }
    const result = await this.documentService.getAttachmentUploadUrl(documentId, user.id);
    this.logger.log(`getDocumentAttachmentUploadUrl: done documentId=${documentId}`);
    this.logger.debug({ uploadUrl: result.uploadUrl }, 'getDocumentAttachmentUploadUrl: response');
    return result;
  }
}
