import { randomUUID } from 'node:crypto';

import { DataSource, type MigrationInterface, type QueryRunner } from 'typeorm';
import { describe, expect, it, vi } from 'vitest';

import { runMigration } from './migration-runner';

describe('migration runner', () => {
  it('logs failure, destroys initialized datasource, shuts down telemetry and returns failure', async () => {
    const datasource = {
      isInitialized: true,
      logger: { logSchemaBuild: vi.fn() },
      initialize: vi.fn().mockResolvedValue(undefined),
      runMigrations: vi.fn().mockRejectedValue(new Error('migration failed')),
      destroy: vi.fn().mockResolvedValue(undefined),
    };
    const sdk = { shutdown: vi.fn().mockResolvedValue(undefined) };
    const logger = { info: vi.fn(), error: vi.fn() };
    await expect(runMigration(datasource as never, sdk, logger)).resolves.toBe(1);
    expect(logger.error).toHaveBeenCalled();
    expect(datasource.destroy).toHaveBeenCalledOnce();
    expect(sdk.shutdown).toHaveBeenCalledOnce();
  });
});

const postgresUrl = process.env.DOCUMENT_TEST_POSTGRES_URL;
it.skipIf(!postgresUrl)(
  'logs the committed migration before a later migration rolls back',
  async () => {
    const schema = `task4_${randomUUID().replaceAll('-', '')}`;
    const inspector = new DataSource({ type: 'postgres', url: postgresUrl });
    await inspector.initialize();
    await inspector.query(`CREATE SCHEMA "${schema}"`);
    const logger = { info: vi.fn(), error: vi.fn() };
    let firstLoggedBeforeSecond = false;
    class First1700000000001 implements MigrationInterface {
      async up(queryRunner: QueryRunner) {
        await queryRunner.query(`CREATE TABLE "${schema}".first_committed (id int)`);
      }
      async down() {}
    }
    class Second1700000000002 implements MigrationInterface {
      async up(queryRunner: QueryRunner) {
        firstLoggedBeforeSecond = logger.info.mock.calls.some(
          ([entry]) => entry.migration === 'First1700000000001'
        );
        await queryRunner.query(`CREATE TABLE "${schema}".second_rolled_back (id int)`);
        throw new Error('deliberate failure');
      }
      async down() {}
    }
    const datasource = new DataSource({
      type: 'postgres',
      url: postgresUrl,
      schema,
      migrations: [First1700000000001, Second1700000000002],
    });
    const sdk = { shutdown: vi.fn().mockResolvedValue(undefined) };
    try {
      expect(await runMigration(datasource, sdk, logger)).toBe(1);
      expect(logger.info).toHaveBeenCalledWith(
        { migration: 'First1700000000001' },
        'Migration applied'
      );
      expect(firstLoggedBeforeSecond).toBe(true);
      const rows = await inspector.query(
        `SELECT tablename FROM pg_tables WHERE schemaname = $1 ORDER BY tablename`,
        [schema]
      );
      expect(rows.map((row: { tablename: string }) => row.tablename)).toEqual([
        'first_committed',
        'migrations',
      ]);
      expect(await inspector.query(`SELECT name FROM "${schema}".migrations`)).toEqual([
        { name: 'First1700000000001' },
      ]);
      expect(sdk.shutdown).toHaveBeenCalledOnce();
    } finally {
      await inspector.query(`DROP SCHEMA "${schema}" CASCADE`);
      await inspector.destroy();
    }
  },
  10000
);
