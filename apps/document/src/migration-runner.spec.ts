import { describe, expect, it, vi } from 'vitest';

import { runMigration } from './migration-runner';

describe('migration runner', () => {
  it('logs failure, destroys initialized datasource, shuts down telemetry and returns failure', async () => {
    const datasource = {
      isInitialized: true,
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
