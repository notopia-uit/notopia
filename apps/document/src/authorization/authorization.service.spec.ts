import { status } from '@grpc/grpc-js';
import { type ClientGrpc } from '@nestjs/microservices';

import { NoteService } from '../note/note.service';
import { AuthorizationService } from './authorization.service';

describe('AuthorizationService.hasNotePermission', () => {
  it('returns false for a non-member instead of failing', async () => {
    const noteService = {
      getWorkspaceByNote: vi
        .fn<NoteService['getWorkspaceByNote']>()
        .mockRejectedValue({ code: status.PERMISSION_DENIED }),
    };
    const service = new AuthorizationService(
      {} as ClientGrpc,
      noteService as unknown as NoteService
    );
    await expect(
      service.hasNotePermission({ memberId: 'u1', documentId: 'd1', permission: 'write' })
    ).resolves.toBe(false);
  });
});
