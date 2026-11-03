import type { DataSource } from 'typeorm';

export async function runMigration(
  datasource: DataSource,
  sdk: { shutdown(): Promise<void> },
  logger: {
    info(object: unknown, message: string): void;
    error(object: unknown, message: string): void;
  }
): Promise<number> {
  let status = 0;
  const typeormLogger = datasource.logger;
  const logSchemaBuild = typeormLogger.logSchemaBuild;
  // TypeORM emits this event after each migration's transaction commits.
  typeormLogger.logSchemaBuild = (message, queryRunner) => {
    const applied = /^Migration (.+) has been executed successfully\.$/.exec(message);
    if (applied) logger.info({ migration: applied[1] }, 'Migration applied');
    logSchemaBuild.call(typeormLogger, message, queryRunner);
  };
  try {
    await datasource.initialize();
    await datasource.runMigrations({ transaction: 'each' });
  } catch (error) {
    status = 1;
    logger.error({ err: error }, 'Migration failed');
  } finally {
    typeormLogger.logSchemaBuild = logSchemaBuild;
    try {
      if (datasource.isInitialized) await datasource.destroy();
    } catch (error) {
      status = 1;
      logger.error({ err: error }, 'Datasource shutdown failed');
    }
    try {
      await sdk.shutdown();
    } catch (error) {
      status = 1;
      logger.error({ err: error }, 'Telemetry shutdown failed');
    }
  }
  return status;
}
