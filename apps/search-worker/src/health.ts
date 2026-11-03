import { Controller, Get, Injectable } from '@nestjs/common';
import { HealthCheck, HealthCheckService, HealthIndicatorService } from '@nestjs/terminus';
import { Meilisearch } from 'meilisearch';

@Injectable()
export class KafkaConsumerState {
  connected = false;
}

@Controller('search-worker/health')
export class HealthController {
  constructor(
    private readonly kafka: KafkaConsumerState,
    private readonly meili: Meilisearch,
    private readonly health: HealthCheckService,
    private readonly indicator: HealthIndicatorService
  ) {}

  @Get('live')
  live() {
    return { status: 'ok' };
  }

  @Get('ready')
  @HealthCheck()
  ready() {
    return this.health.check([
      () =>
        this.kafka.connected
          ? this.indicator.check('kafka').up()
          : this.indicator.check('kafka').down(),
      async () => {
        try {
          await this.meili.health();
          return this.indicator.check('meilisearch').up();
        } catch {
          return this.indicator.check('meilisearch').down();
        }
      },
    ]);
  }
}
