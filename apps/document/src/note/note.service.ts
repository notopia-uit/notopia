import { Inject, Injectable, OnModuleInit } from '@nestjs/common';
import { type ClientGrpc } from '@nestjs/microservices';
import {
  NOTE_PACKAGE_NAME,
  NOTE_SERVICE_NAME,
  NoteServiceClient,
} from '@notopia-uit/pb/note';
import { firstValueFrom } from 'rxjs';

@Injectable()
export class NoteService implements OnModuleInit {
  private noteServiceClient!: NoteServiceClient;

  constructor(@Inject(NOTE_PACKAGE_NAME) private client: ClientGrpc) {}

  onModuleInit(): void {
    this.noteServiceClient =
      this.client.getService<NoteServiceClient>(NOTE_SERVICE_NAME);
  }

  // TODO: previously used to fetch each, but we might going to change batch, maybe this will be removed
  async getNoteName(noteId: string): Promise<string> {
    const response = await firstValueFrom(
      this.noteServiceClient.getNoteName({ id: noteId })
    );
    return response.name;
  }

  async checkNoteExistence(noteId: string): Promise<boolean> {
    const response = await firstValueFrom(
      this.noteServiceClient.checkNoteExistence({
        noteId,
      })
    );
    return response.exists;
  }

  async getWorkspaceIdByNoteId(noteId: string): Promise<string> {
    const response = await firstValueFrom(
      this.noteServiceClient.getWorkspaceIdByNoteId({
        noteId,
      })
    );
    return response.workspaceId;
  }
}
