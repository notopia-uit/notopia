import { status } from '@grpc/grpc-js';
import { Inject, Injectable, InternalServerErrorException, Logger, OnModuleInit } from '@nestjs/common';
import { type ClientGrpc } from '@nestjs/microservices';
import {
  NOTE_PACKAGE_NAME,
  NOTE_SERVICE_NAME,
  Note,
  NoteServiceClient,
  Trashed,
  TrashedBy,
  Workspace,
} from '@notopia-uit/pb/note';
import { firstValueFrom } from 'rxjs';

import { isGrpcError, protoTimestampToDate } from '#/common/grpc';
import { NoteNotFoundException } from '#/note/note-not-found.exception';
import { WorkspaceNoteNotFoundException } from '#/note/workspace-note-not-found.exception';

import { NoteModel, TrashedModel, WorkspaceModel } from './models';

@Injectable()
export class NoteService implements OnModuleInit {
  private readonly logger = new Logger(NoteService.name);
  private noteServiceClient!: NoteServiceClient;

  constructor(@Inject(NOTE_PACKAGE_NAME) private readonly client: ClientGrpc) {}

  onModuleInit(): void {
    this.noteServiceClient = this.client.getService<NoteServiceClient>(NOTE_SERVICE_NAME);
  }

  async getNoteById({
    noteId,
    userId,
    excludeTrashed = false,
  }: {
    noteId: string;
    userId: string;
    excludeTrashed?: boolean;
  }): Promise<NoteModel> {
    this.logger.debug(`getNoteById: noteId=${noteId} userId=${userId}`);
    let note: Note | undefined;
    try {
      const response = await firstValueFrom(
        this.noteServiceClient.getNote({
          id: noteId,
          userId: userId,
          excludeTrashed: excludeTrashed,
        })
      );
      note = response.note;
    } catch (error) {
      if (isGrpcError(error) && error.code === status.NOT_FOUND) {
        this.logger.warn(`getNoteById: not found noteId=${noteId}`);
        throw new NoteNotFoundException(noteId);
      }
      this.logger.warn(`getNoteById: gRPC error noteId=${noteId}`);
      throw error;
    }
    if (!note) {
      this.logger.warn(`getNoteById: empty response noteId=${noteId}`);
      throw new NoteNotFoundException(noteId);
    }

    if (!note.updatedAt) {
      throw new InternalServerErrorException(
        `Note with id ${noteId} is missing updatedAt timestamp`
      );
    }
    return {
      id: note.id,
      name: note.name,
      tags: note.tags,
      folderId: note.folderId,
      icon: note.icon,
      updatedAt: protoTimestampToDate(note.updatedAt),
      trashed: note.trashed ? this.toTrashedModel(note.trashed) : undefined,
    };
  }

  // TODO: previously used to fetch each, but we might going to change batch, maybe this will be removed
  async getNoteName(noteId: string): Promise<string> {
    this.logger.debug(`getNoteName: noteId=${noteId}`);
    const response = await firstValueFrom(this.noteServiceClient.getNoteName({ id: noteId }));
    return response.name;
  }

  async getWorkspaceByNote({
    userId,
    noteId,
  }: {
    userId: string;
    noteId: string;
  }): Promise<WorkspaceModel> {
    this.logger.debug(`getWorkspaceByNote: noteId=${noteId} userId=${userId}`);
    let workspace: Workspace | undefined;
    try {
      const response = await firstValueFrom(
        this.noteServiceClient.getWorkspaceByNote({
          noteId: noteId,
          userId: userId,
        })
      );
      workspace = response.workspace;
    } catch (error) {
      if (isGrpcError(error) && error.code === status.NOT_FOUND) {
        this.logger.warn(`getWorkspaceByNote: not found noteId=${noteId}`);
        throw new WorkspaceNoteNotFoundException(noteId);
      }
      this.logger.warn(`getWorkspaceByNote: gRPC error noteId=${noteId}`);
      throw error;
    }
    if (!workspace) {
      this.logger.warn(`getWorkspaceByNote: empty response noteId=${noteId}`);
      throw new WorkspaceNoteNotFoundException(noteId);
    }
    return {
      id: workspace.id,
      name: workspace.name,
      slug: workspace.slug,
    };
  }

  toTrashedModel(trashed: Trashed): TrashedModel {
    switch (trashed.by) {
      case TrashedBy.TRASHED_BY_UNSPECIFIED:
        throw new InternalServerErrorException('Trashed by value is unspecified');
      case TrashedBy.TRASHED_BY_PURPOSE:
        if (!trashed.at) {
          throw new InternalServerErrorException(
            'Trashed at timestamp is required for TRASHED_BY_PURPOSE'
          );
        }
        return {
          by: 'purpose',
          at: protoTimestampToDate(trashed.at),
        };
      case TrashedBy.TRASHED_BY_PARENT:
        if (!trashed.at) {
          throw new InternalServerErrorException(
            'Trashed at timestamp is required for TRASHED_BY_PARENT'
          );
        }
        return {
          by: 'parent',
          at: protoTimestampToDate(trashed.at),
        };
      default:
        throw new InternalServerErrorException('Unknown trashed by value');
    }
  }
}
