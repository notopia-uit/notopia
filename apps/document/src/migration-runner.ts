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
  try {
    await datasource.initialize();
    const applied = await datasource.runMigrations({ transaction: 'each' });
    for (const migration of applied)
      logger.info({ migration: migration.name }, 'Migration applied');
  } catch (error) {
    status = 1;
    logger.error({ err: error }, 'Migration failed');
  } finally {
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
