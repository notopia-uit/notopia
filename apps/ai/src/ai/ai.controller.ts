import {
  Body,
  Controller,
  Headers,
  Inject,
  Post,
  Res,
  UseFilters,
  UseGuards,
} from '@nestjs/common';
import type { ConfigType } from '@nestjs/config';
import { pipeUIMessageStreamToResponse } from 'ai';
import type { Response } from 'express';

import { HttpUserGuard, ReqUser, type User } from '#/common';
import { AiError, AiExceptionFilter } from '#/common/ai-error';
import { usageConfig } from '#/config.factory';

import { AiService } from './ai.service';
import { parseChatRequest } from './chat-request.dto';

@Controller('ai')
@UseFilters(AiExceptionFilter)
export class AiController {
  constructor(
    private readonly aiService: AiService,
    @Inject(usageConfig.KEY) private readonly config: ConfigType<typeof usageConfig>
  ) {}

  @Post()
  @UseGuards(HttpUserGuard)
  async chat(
    @Body() input: unknown,
    @ReqUser() user: User,
    @Res() res: Response,
    @Headers('x-workspace-id') workspaceId?: string
  ): Promise<void> {
    const abort = new AbortController();
    const timer = setTimeout(
      () => abort.abort(new AiError(504, 'aiRequestTimeout', 'The AI request timed out.')),
      this.config.timeoutMs
    );
    const close = () => {
      clearTimeout(timer);
      abort.abort();
    };
    res.once('close', close);
    res.once('finish', close);
    try {
      const body = await parseChatRequest(input);
      const stream = await this.aiService.streamBlockNoteChat(
        body,
        user.id,
        abort.signal,
        workspaceId
      );
      if (abort.signal.aborted) {
        await stream.cancel();
        return;
      }
      await pipeUIMessageStreamToResponse({ response: res, stream });
    } finally {
      clearTimeout(timer);
      res.off('close', close);
      res.off('finish', close);
      abort.abort();
    }
  }
}
