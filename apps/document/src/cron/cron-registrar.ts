import { Injectable, Logger, OnModuleInit } from '@nestjs/common';
import { ConfigService } from '@nestjs/config';
import { LockNotAcquiredError, Locks } from '@nestjs/locks';
import { SchedulerRegistry } from '@nestjs/schedule';
import { CronJob } from 'cron';

import { CRON_CONFIG, CronConfig, CronJobConfig } from '../config';
import { RevisionPurgeCron } from './revision-purge.cron';
import { SweeperCron } from './sweeper.cron';

type CronRun = (signal: AbortSignal) => Promise<void>;

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
    this.register(this.cronConfig.sweeper, (signal) => this.sweeper.sweep(signal));
    this.register(this.cronConfig.revisionPurge, (signal) => this.revisionPurge.purge(signal));
  }

  private register(cfg: CronJobConfig, run: CronRun): void {
    if (!cfg.enabled) {
      this.logger.log({ job: cfg.name }, 'cron job disabled');
      return;
    }
    // No timeZone: the job runs in the process zone (TZ env, else the system zone).
    const job = CronJob.from({
      cronTime: cfg.schedule,
      onTick: () => this.runLocked(cfg, run),
    });
    this.schedulerRegistry.addCronJob(cfg.name, job);
    job.start();
    this.logger.log(
      { job: cfg.name, schedule: cfg.schedule, timeoutSeconds: cfg.timeoutSeconds },
      'cron job registered'
    );
  }

  private async runLocked(cfg: CronJobConfig, run: CronRun): Promise<void> {
    const name = cfg.name;
    this.logger.debug({ job: name }, 'cron job tick');
    try {
      await this.locks.withLock(name, async (lock) => {
        const timeout = AbortSignal.timeout(cfg.timeoutSeconds * 1000);
        await run(AbortSignal.any([lock.signal, timeout]));
        if (timeout.aborted) {
          this.logger.warn({ job: name, timeoutSeconds: cfg.timeoutSeconds }, 'cron job timed out');
        }
      });
    } catch (err) {
      if (err instanceof LockNotAcquiredError) {
        this.logger.debug({ job: name }, 'cron job held by another instance, skipping');
        return;
      }
      this.logger.error({ err, job: name }, 'cron job failed');
    }
  }
}
