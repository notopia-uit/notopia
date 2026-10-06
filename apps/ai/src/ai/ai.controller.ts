import { Body, Controller, Post, Res, UseGuards } from '@nestjs/common';
import { pipeUIMessageStreamToResponse } from 'ai';
import type { Response } from 'express';
import { Traceable } from 'nestjs-otel';

import { HttpUserGuard, ReqUser, type User } from '#/common';

import { AiService } from './ai.service';
import { BlockNoteChatBody } from './chat-request.dto';

@Controller('ai')
@Traceable()
export class AiController {
  constructor(private readonly aiService: AiService) {}

  @Post()
  @UseGuards(HttpUserGuard)
  async chat(
    @Body() body: BlockNoteChatBody,
    @ReqUser() user: User,
    @Res() res: Response
  ): Promise<void> {
    const stream = await this.aiService.streamBlockNoteChat(body, user.id);
    return pipeUIMessageStreamToResponse({ response: res, stream });
  }
}
