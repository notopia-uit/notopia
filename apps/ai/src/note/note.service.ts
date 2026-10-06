import { status } from '@grpc/grpc-js';
import { Inject, Injectable, Logger, OnModuleInit } from '@nestjs/common';
import type { ClientGrpc } from '@nestjs/microservices';
import { NOTE_PACKAGE_NAME, NOTE_SERVICE_NAME, type NoteServiceClient } from '@notopia-uit/pb/note';
import { firstValueFrom } from 'rxjs';

import { isGrpcError } from '#/common';

export type NoteSummary = {
  id: string;
  name: string;
  tags: string[];
  folderId?: string;
};

@Injectable()
export class NoteService implements OnModuleInit {
  private readonly logger = new Logger(NoteService.name);
  private client!: NoteServiceClient;

  constructor(@Inject(NOTE_PACKAGE_NAME) private readonly grpc: ClientGrpc) {}

  onModuleInit(): void {
    this.client = this.grpc.getService<NoteServiceClient>(NOTE_SERVICE_NAME);
  }

  async getNote({
    userId,
    noteId,
  }: {
    userId: string;
    noteId: string;
  }): Promise<NoteSummary | null> {
    this.logger.debug({ noteId }, 'Fetching note for tool call');
    try {
      const response = await firstValueFrom(
        this.client.getNote({ id: noteId, userId, excludeTrashed: true })
      );
      const note = response.note;
      if (!note) {
        return null;
      }
      return {
        id: note.id,
        name: note.name,
        tags: note.tags,
        folderId: note.folderId,
      };
    } catch (error) {
      if (isGrpcError(error) && error.code === status.NOT_FOUND) {
        return null;
      }
      throw error;
    }
  }
}
