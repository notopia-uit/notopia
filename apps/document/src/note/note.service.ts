import { Client, Code, ConnectError } from '@connectrpc/connect';
import {
  Injectable,
  InternalServerErrorException,
  NotFoundException,
} from '@nestjs/common';
import { NoteService as NoteServiceDefinition } from '@notopia-uit/pb/note';

export type NoteClient = Client<typeof NoteServiceDefinition>;

@Injectable()
export class NoteService {
  constructor(private readonly noteClient: NoteClient) {}

  async getNoteName(noteId: string): Promise<string> {
    try {
      const response = await this.noteClient.getNoteName({ id: noteId });
      return response.name;
    } catch (error) {
      if (error instanceof ConnectError) {
        if (error.code === Code.NotFound) {
          throw new NotFoundException(`Note with ID ${noteId} not found`);
        }
        throw new InternalServerErrorException(
          `Failed to get note name: ${error.message}`
        );
      }
      throw error;
    }
  }

  async checkNoteExistence(noteId: string): Promise<boolean> {
    try {
      const response = await this.noteClient.checkNoteExistence({ noteId });
      return response.exists;
    } catch (error) {
      if (error instanceof ConnectError) {
        if (error.code === Code.NotFound) {
          return false;
        }
        throw new InternalServerErrorException(
          `Failed to check note existence: ${error.message}`
        );
      }
      throw error;
    }
  }

  async getWorkspaceIdByNoteId(noteId: string): Promise<string> {
    try {
      const response = await this.noteClient.getWorkspaceIdByNoteId({ noteId });
      return response.workspaceId;
    } catch (error) {
      if (error instanceof ConnectError) {
        if (error.code === Code.NotFound) {
          throw new NotFoundException(`Note with ID ${noteId} not found`);
        }
        throw new InternalServerErrorException(
          `Failed to get workspace ID by note ID: ${error.message}`
        );
      }
      throw error;
    }
  }
}
