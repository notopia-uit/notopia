import { HttpStatus, Logger, UnauthorizedException } from '@nestjs/common';
import { DocumentApi as DocumentApiDefinition } from '@notopia-uit/api-document-nestjs-server/api';
import { CommittedRevision } from '@notopia-uit/api-document-nestjs-server/models';
import type { Request as ExpressRequest } from 'express';
import { Traceable } from 'nestjs-otel';

import { User } from '#/common';

import { AuthorizationService } from '../authorization/authorization.service';
import { HocuspocusService } from '../hocuspocus';
import { DocumentPermissionException } from './document-permission.exception';
import { DocumentService } from './document.service';

@Traceable()
export class DocumentApi extends DocumentApiDefinition {
  private readonly logger = new Logger(DocumentApi.name);

  constructor(
    private readonly documentService: DocumentService,
    private readonly hocuspocusService: HocuspocusService,
    private readonly authorizationService: AuthorizationService
  ) {
    super();
  }

  override async commitDocument(documentId: string, req: Request): Promise<CommittedRevision> {
    this.logger.log({ documentId }, 'commitDocument: received');
    const user = (req as unknown as Record<string, unknown>).user as User | undefined;
    if (!user) {
      throw new UnauthorizedException('User not authenticated');
    }
    const res = (req as unknown as ExpressRequest).res;
    if (!res) {
      throw new Error('commitDocument: express response not available on the request');
    }
    try {
      const canWrite = await this.authorizationService.hasNotePermission({
        documentId,
        memberId: user.id,
        permission: 'write',
      });
      if (!canWrite) {
        throw new DocumentPermissionException(documentId, user.id);
      }
      const { revisionId, modifiedCleared, committedContentHash } =
        await this.documentService.commitIfDue({
          documentId,
          trigger: 'manual',
          authorId: user.id,
          force: true,
          minIntervalSeconds: 0,
        });
      if (modifiedCleared && committedContentHash) {
        this.hocuspocusService.clearModifiedIfCommitted(documentId, committedContentHash);
      }
      // The generated controller has no per-response status, but Nest sets the route's default
      // status (201 for POST) before calling the handler and does not set it again when
      // replying, so the status set on `res` wins:
      // - 201 {id}: a new revision was created
      // - 200 {id}: nothing changed (not modified, or same content as the latest revision);
      //   id is the latest revision
      // - 204: nothing changed and the document has no revision yet
      if (revisionId) {
        res.status(HttpStatus.CREATED);
        this.logger.log({ documentId, revisionId }, 'commitDocument: committed');
        return { id: revisionId };
      }
      const latestRevisionId = await this.documentService.getLatestRevisionId(documentId);
      if (!latestRevisionId) {
        res.status(HttpStatus.NO_CONTENT);
        this.logger.log({ documentId }, 'commitDocument: nothing to commit, no revision');
        // 204 has no body; the generated signature has no void variant
        return undefined as unknown as CommittedRevision;
      }
      res.status(HttpStatus.OK);
      const response: CommittedRevision = { id: latestRevisionId };
      this.logger.log({ documentId }, 'commitDocument: nothing to commit');
      this.logger.debug({ response }, 'commitDocument: response');
      return response;
    } catch (error) {
      this.logger.error({ err: error, documentId }, 'commitDocument: error');
      throw error;
    }
  }

  override async getDocumentAttachmentUploadUrl(
    documentId: string,
    filename: string,
    req: Request
  ) {
    this.logger.log({ documentId, filename }, 'getDocumentAttachmentUploadUrl: received');
    const user = (req as unknown as Record<string, unknown>).user as User | undefined;
    if (!user) {
      throw new UnauthorizedException('User not authenticated');
    }
    try {
      const result = await this.documentService.getAttachmentUploadUrl({
        documentId,
        filename,
        userId: user.id,
      });
      this.logger.log({ documentId, filename }, 'getDocumentAttachmentUploadUrl: done');
      this.logger.debug(
        { uploadUrl: result.uploadUrl },
        'getDocumentAttachmentUploadUrl: response'
      );
      return result;
    } catch (error) {
      this.logger.error(
        { err: error, documentId, filename },
        'getDocumentAttachmentUploadUrl: error'
      );
      throw error;
    }
  }
}
