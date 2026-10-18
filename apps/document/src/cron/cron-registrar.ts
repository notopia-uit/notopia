import { Injectable, Logger, OnModuleInit } from '@nestjs/common';
import { ConfigService } from '@nestjs/config';
import { LockNotAcquiredError, Locks } from '@nestjs/locks';
import { SchedulerRegistry } from '@nestjs/schedule';
import { CronJob } from 'cron';

import { CRON_CONFIG, CronConfig, CronJobConfig } from '../config';
import { RevisionPurgeCron } from './revision-purge.cron';
import { SweeperCron } from './sweeper.cron';

// NOTE: move to a document worker when crons are split out of the API process.
@Injectable()
export class CronRegistrar implements OnModuleInit {
  private readonly logger = new Logger(CronRegistrar.name);
  private readonly cronConfig: CronConfig;

  constructor(
    private readonly schedulerRegistry: SchedulerRegistry,
    private readonly locks: Locks,
    private readonly sweeper: SweeperCron,
    private readonly revisionPurge: RevisionPurgeCron,
    configService: ConfigService
  ) {
    const cronConfig = configService.get<CronConfig>(CRON_CONFIG);
    if (!cronConfig) {
      throw new Error('CRON_CONFIG not found');
    }
    this.cronConfig = cronConfig;
  }

  onModuleInit(): void {
    this.register(this.cronConfig.sweeper, () => this.sweeper.sweep());
    this.register(this.cronConfig.revisionPurge, () => this.revisionPurge.purge());
  }

  private register(cfg: CronJobConfig, run: () => Promise<void>): void {
    if (!cfg.enabled) {
      this.logger.log({ job: cfg.name }, 'cron job disabled');
      return;
    }
    // No timeZone: the job runs in the process zone (TZ env, else the system zone).
    const job = CronJob.from({
      cronTime: cfg.schedule,
      onTick: () => this.runLocked(cfg.name, run),
    });
    this.schedulerRegistry.addCronJob(cfg.name, job);
    job.start();
    this.logger.log({ job: cfg.name, schedule: cfg.schedule }, 'cron job registered');
  }

  private async runLocked(name: string, run: () => Promise<void>): Promise<void> {
    this.logger.debug({ job: name }, 'cron job tick');
    try {
      await this.locks.withLock(name, () => run());
    } catch (err) {
      if (err instanceof LockNotAcquiredError) {
        this.logger.debug({ job: name }, 'cron job held by another instance, skipping');
        return;
      }
      this.logger.error({ err, job: name }, 'cron job failed');
    }
  }
}
