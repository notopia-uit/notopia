import { Client, Code, ConnectError } from '@connectrpc/connect';
import { Injectable, InternalServerErrorException } from '@nestjs/common';
import { NoteService as NoteServiceDefinition } from '@notopia-uit/pb/note';

export type NoteClient = Client<typeof NoteServiceDefinition>;

@Injectable()
export class NoteService {
  constructor(private readonly noteClient: NoteClient) {}

  async checkNoteExistence(noteId: string): Promise<{ exists: boolean }> {
    try {
      return await this.noteClient.checkNoteExistence({ noteId });
    } catch (error) {
      if (error instanceof ConnectError) {
        if (error.code === Code.NotFound) {
          return { exists: false };
        }
        throw new InternalServerErrorException(
          `Failed to check note existence: ${error.message}`
        );
      }
      throw error;
    }
  }
}
