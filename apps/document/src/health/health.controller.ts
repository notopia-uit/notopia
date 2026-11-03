import { Controller, Get, Inject } from '@nestjs/common';
import { HealthCheck, HealthCheckService } from '@nestjs/terminus';
import { DataSource } from 'typeorm';

import { SCALING_CONFIG, type ScalingConfig } from '../config';

export const HEALTH_REDIS = Symbol('HEALTH_REDIS');

@Controller('document/health')
export class HealthController {
  constructor(
    private readonly health: HealthCheckService,
    private readonly dataSource: DataSource,
    @Inject(SCALING_CONFIG) private readonly scaling: ScalingConfig,
    @Inject(HEALTH_REDIS) private readonly redis: { ping(): Promise<unknown> }
  ) {}

  @Get('live')
  @HealthCheck()
  live() {
    return this.health.check([]);
  }

  @Get('ready')
  @HealthCheck()
  ready() {
    return this.health.check([
      async () => {
        try {
          await this.dataSource.query('SELECT 1');
          return { postgres: { status: 'up' as const } };
        } catch (error) {
          return { postgres: { status: 'down' as const, message: String(error) } };
        }
      },
      ...(this.scaling.enabled
        ? [
            async () => {
              try {
                await this.redis.ping();
                return { redis: { status: 'up' as const } };
              } catch (error) {
                return { redis: { status: 'down' as const, message: String(error) } };
              }
            },
          ]
        : []),
    ]);
  }

  @Get('startup')
  @HealthCheck()
  startup() {
    return this.health.check([
      async () => {
        try {
          return {
            migrations: {
              status: (await this.dataSource.showMigrations())
                ? ('down' as const)
                : ('up' as const),
            },
          };
        } catch (error) {
          return { migrations: { status: 'down' as const, message: String(error) } };
        }
      },
    ]);
  }
}
