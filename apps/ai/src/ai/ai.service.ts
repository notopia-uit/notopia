import { createOpenAICompatible } from '@ai-sdk/openai-compatible';
import {
  aiDocumentFormats,
  injectDocumentStateMessages,
  toolDefinitionsToToolSet,
} from '@blocknote/xl-ai/server';
import { Inject, Injectable } from '@nestjs/common';
import type { ConfigType } from '@nestjs/config';
import {
  convertToModelMessages,
  stepCountIs,
  streamText,
  type UIMessage,
  type UIMessageChunk,
} from 'ai';
import { Traceable } from 'nestjs-otel';

import { AuthorizationService } from '#/authorization/authorization.service';
import { llmConfig } from '#/config.factory';
import { NoteService } from '#/note/note.service';
import { SearchService } from '#/search/search.service';

import { createCustomTools } from './tools/custom-tools';

export type BlockNoteChatRequest = {
  messages: UIMessage[];
  toolDefinitions?: Parameters<typeof toolDefinitionsToToolSet>[0];
};

@Injectable()
@Traceable()
export class AiService {
  constructor(
    @Inject(llmConfig.KEY) private readonly llm: ConfigType<typeof llmConfig>,
    private readonly noteClient: NoteService,
    private readonly search: SearchService,
    private readonly authorization: AuthorizationService
  ) {}

  async streamBlockNoteChat(
    body: BlockNoteChatRequest,
    userId: string
  ): Promise<ReadableStream<UIMessageChunk>> {
    const model = createOpenAICompatible({
      baseURL: this.llm.baseUrl,
      apiKey: this.llm.apiKey,
      name: 'llm',
    })(this.llm.model);

    const result = streamText({
      model,
      system: aiDocumentFormats.html.systemPrompt,
      messages: await convertToModelMessages(injectDocumentStateMessages(body.messages)),
      tools: {
        ...toolDefinitionsToToolSet(body.toolDefinitions ?? {}),
        ...createCustomTools({
          userId,
          noteClient: this.noteClient,
          search: this.search,
          authorization: this.authorization,
        }),
      },
      stopWhen: stepCountIs(10), // TODO: Parameterise this
    });

    return result.toUIMessageStream();
  }
}
