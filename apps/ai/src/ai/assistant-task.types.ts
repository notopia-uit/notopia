import type { UIMessage } from 'ai';

export type TaskOwner = { userId: string; workspaceId: string };
export type PendingClientCall = {
  toolCallId: string;
  toolName: string;
  input: unknown;
};
export type TaskOutcome = 'completed' | 'cancelled' | 'failed';
export type AssistantTask = {
  id: string;
  owner: TaskOwner;
  promptId: string;
  messages: UIMessage[];
  deadline: number;
  remainingSteps: number;
  targetNoteId?: string;
  pendingCalls: PendingClientCall[];
  state: 'running' | 'awaiting-client' | TaskOutcome;
  revision: number;
};
