import { isToolUIPart, getToolName, type UIMessage } from 'ai';

export interface AssistantTaskData {
  taskId: string;
  expiresAt: number;
}
export interface AssistantContinuation extends AssistantTaskData {
  token: string;
}
export interface AssistantResult {
  toolCallId: string;
  toolName: 'openNoteForEditing' | 'applyDocumentOperations';
  output: unknown;
}

export function completedAssistantResults(message: UIMessage | undefined): AssistantResult[] {
  return (message?.parts ?? []).flatMap((part) => {
    if (!isToolUIPart(part)) return [];
    const name = getToolName(part);
    if (name !== 'openNoteForEditing' && name !== 'applyDocumentOperations') return [];
    if (part.state !== 'output-available')
      throw new Error('The editor could not complete the assistant operation.');
    return [{ toolCallId: part.toolCallId, toolName: name, output: part.output }];
  });
}
