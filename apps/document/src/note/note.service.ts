import { Inject, Injectable, OnModuleInit } from '@nestjs/common';
import { type ClientGrpc } from '@nestjs/microservices';
import {
  NOTE_PACKAGE_NAME,
  NOTE_SERVICE_NAME,
  NoteServiceClient,
  Trashed,
  TrashedBy,
} from '@notopia-uit/pb/note';
import { firstValueFrom } from 'rxjs';

import { protoTimestampToDate } from '#/common/proto-timestamp';

import { NoteModel, TrashedModel } from './note.model';

@Injectable()
export class NoteService implements OnModuleInit {
  private noteServiceClient!: NoteServiceClient;

  constructor(@Inject(NOTE_PACKAGE_NAME) private client: ClientGrpc) {}

  onModuleInit(): void {
    this.noteServiceClient =
      this.client.getService<NoteServiceClient>(NOTE_SERVICE_NAME);
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
    const response = await firstValueFrom(
      this.noteServiceClient.getNote({
        id: noteId,
        userId: userId,
        excludeTrashed: excludeTrashed,
      })
    );
    if (!response.updatedAt) {
      throw new Error('Note updatedAt timestamp is required');
    }
    return {
      id: response.id,
      name: response.name,
      tags: response.tags,
      folderId: response.folderId,
      icon: response.icon,
      updatedAt: protoTimestampToDate(response.updatedAt),
      trashed: response.trashed
        ? this.toTrashedModel(response.trashed)
        : undefined,
    };
  }

  // TODO: previously used to fetch each, but we might going to change batch, maybe this will be removed
  async getNoteName(noteId: string): Promise<string> {
    const response = await firstValueFrom(
      this.noteServiceClient.getNoteName({ id: noteId })
    );
    return response.name;
  }

  // TODO: Get note hey

  async getWorkspaceIdByNoteId(noteId: string): Promise<string> {
    const response = await firstValueFrom(
      this.noteServiceClient.getWorkspaceIdByNoteId({
        noteId,
      })
    );
    return response.workspaceId;
  }

  toTrashedModel(trashed: Trashed): TrashedModel {
    // export enum TrashedBy {
    //   TRASHED_BY_UNSPECIFIED = 0,
    //   TRASHED_BY_PURPOSE = 1,
    //   TRASHED_BY_PARENT = 2,
    //   UNRECOGNIZED = -1,
    // }
    switch (trashed.by) {
      case TrashedBy.TRASHED_BY_UNSPECIFIED:
        throw new Error('Invalid trashed by value: TRASHED_BY_UNSPECIFIED');
      case TrashedBy.TRASHED_BY_PURPOSE:
        if (!trashed.at) {
          throw new Error(
            'Trashed at timestamp is required for TRASHED_BY_PURPOSE'
          );
        }
        return {
          by: 'purpose',
          at: protoTimestampToDate(trashed.at),
        };
      case TrashedBy.TRASHED_BY_PARENT:
        if (!trashed.at) {
          throw new Error(
            'Trashed at timestamp is required for TRASHED_BY_PARENT'
          );
        }
        return {
          by: 'parent',
          at: protoTimestampToDate(trashed.at),
        };
      default:
        throw new Error(`Invalid trashed by value: ${trashed.by}`);
    }
  }
}
