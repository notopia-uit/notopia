import { Controller, Get } from '@nestjs/common';

@Controller()
export class HealthController {
  @Get('ai/health')
  health(): { status: string } {
    return { status: 'ok' };
  }
}
