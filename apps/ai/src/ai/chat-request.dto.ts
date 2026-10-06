import type { UIMessage } from 'ai';

import type { BlockNoteChatRequest } from './ai.service';

export class BlockNoteChatBody implements BlockNoteChatRequest {
  messages!: UIMessage[];
  toolDefinitions?: BlockNoteChatRequest['toolDefinitions'];
}
