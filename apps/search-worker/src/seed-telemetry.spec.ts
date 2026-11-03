import { spawn } from 'node:child_process';
import { once } from 'node:events';
import { createServer } from 'node:http';

import { expect, it, vi } from 'vitest';

it('exports the production seed span without credentials even when instrumentation is enabled', async () => {
  const master = 'test-master-secret-telemetry';
  const searchKey = 'test-search-secret-telemetry';
  const server = createServer((request, response) => {
    response.setHeader('content-type', 'application/json');
    if (request.url?.endsWith('/settings')) {
      response.end(
        JSON.stringify({
          searchableAttributes: ['name', 'folderName', 'plainTextContent', 'tags'],
          filterableAttributes: ['workspaceId', 'tags'],
          sortableAttributes: ['name'],
        })
      );
    } else if (request.url?.includes('/keys/')) {
      response.writeHead(404);
      response.end(JSON.stringify({ message: 'not found', code: 'api_key_not_found' }));
    } else if (request.url === '/keys') {
      response.end(JSON.stringify({ uid: '00000000-0000-4000-8000-000000000001', key: searchKey }));
    } else {
      response.end(JSON.stringify({ uid: 'notes', primaryKey: 'id' }));
    }
  });
  server.listen(0, '127.0.0.1');
  await once(server, 'listening');
  const address = server.address();
  if (!address || typeof address === 'string') throw new Error('Missing listener address');
  try {
    const child = spawn('pnpm', ['exec', 'tsx', 'seed/main.ts'], {
      cwd: import.meta.dirname + '/..',
      env: {
        ...process.env,
        NODE_ENV: 'production',
        OTEL_SDK_DISABLED: 'false',
        OTEL_TRACES_EXPORTER: 'console',
        OTEL_METRICS_EXPORTER: 'none',
        OTEL_LOGS_EXPORTER: 'none',
        OTEL_NODE_ENABLED_INSTRUMENTATIONS: 'http,undici,pino',
        NOTOPIA_SEARCH_WORKER_MEILI_HOST: `http://127.0.0.1:${address.port}`,
        NOTOPIA_SEARCH_WORKER_MEILI_API_KEY: master,
        NOTOPIA_SEARCH_WORKER_MEILI_NOTE_SEARCH_KEY_UID: '00000000-0000-4000-8000-000000000001',
      },
    });
    let output = '';
    child.stdout.on('data', (chunk: Buffer) => {
      output += chunk.toString();
    });
    child.stderr.on('data', (chunk: Buffer) => {
      output += chunk.toString();
    });
    const [code] = await once(child, 'close');
    expect(code, output.replaceAll(master, '[redacted]').replaceAll(searchKey, '[redacted]')).toBe(
      0
    );
    expect(output).toContain('cli.seed');
    expect(output).not.toContain(master);
    expect(output).not.toContain(searchKey);
  } finally {
    await new Promise<void>((resolve, reject) =>
      server.close((error) => (error ? reject(error) : resolve()))
    );
  }
}, 20000);

it('does not expose telemetry shutdown errors on SIGTERM', async () => {
  const secret = 'shutdown-secret';
  let signalHandler: (() => void) | undefined;
  const originalOnce = process.once.bind(process);
  const originalExitCode = process.exitCode;
  vi.spyOn(process, 'once').mockImplementation(((event: string, listener: () => void) => {
    if (event === 'SIGTERM') {
      signalHandler = listener;
      return process;
    }
    return originalOnce(event, listener);
  }) as typeof process.once);
  const stderr = vi.spyOn(process.stderr, 'write').mockImplementation(() => true);
  const exit = vi.spyOn(process, 'exit').mockImplementation((() => undefined) as never);
  vi.doMock('@opentelemetry/sdk-node', () => ({
    NodeSDK: class {
      start() {}
      shutdown() {
        return Promise.reject(new Error(secret));
      }
    },
  }));
  try {
    await import('../seed/otel');
    expect(signalHandler).toBeDefined();
    signalHandler?.();
    await vi.waitFor(() => expect(exit).toHaveBeenCalledWith(1));
    const output = stderr.mock.calls.flat().join('');
    expect(output).toContain('Failed to shut down telemetry');
    expect(output).not.toContain(secret);
  } finally {
    process.exitCode = originalExitCode;
    vi.restoreAllMocks();
    vi.doUnmock('@opentelemetry/sdk-node');
    vi.resetModules();
  }
});
