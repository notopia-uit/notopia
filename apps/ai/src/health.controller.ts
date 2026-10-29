import { Controller, Get, Res } from '@nestjs/common';
import type { Response } from 'express';

import { UsageService } from './usage/usage.service';

interface HealthResponse {
  status: 'up' | 'down';
  details?: Record<string, { status: 'up' | 'down' }>;
}

@Controller('ai/health')
export class HealthController {
  constructor(private readonly usage: UsageService) {}

  @Get()
  health(): HealthResponse {
    return this.live();
  }

  @Get('live')
  live(): HealthResponse {
    return { status: 'up' };
  }

  @Get('startup')
  startup(): HealthResponse {
    return { status: 'up' };
  }

  @Get('ready')
  async ready(@Res({ passthrough: true }) response: Response): Promise<HealthResponse> {
    try {
      await this.usage.pingRedis();
      return { status: 'up', details: { redis: { status: 'up' } } };
    } catch {
      response.status(503);
      return { status: 'down', details: { redis: { status: 'down' } } };
    }
  }
}
