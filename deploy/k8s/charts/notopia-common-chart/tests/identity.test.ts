import assert from 'node:assert/strict';
import { spawnSync } from 'node:child_process';
import { join } from 'node:path';
import { before, test } from 'node:test';

const charts = join(import.meta.dirname, '../..');
const services = ['authorization', 'billing', 'document', 'note', 'search-worker', 'web'];

before(() => {
  for (const chart of services) {
    const result = spawnSync(
      'flock',
      [
        `tmp/chart-lock/${chart}.lock`,
        'helm',
        'dependency',
        'build',
        '--skip-refresh',
        join(charts, chart),
      ],
      { encoding: 'utf8' }
    );
    assert.equal(result.status, 0, result.stderr);
  }
});

for (const chart of services) {
  for (const field of ['runAsUser', 'runAsGroup', 'fsGroup']) {
    test(`${chart} rejects null ${field}`, () => {
      const args = [
        'template',
        'test',
        join(charts, chart),
        '-f',
        '-',
        '--set',
        'istio.enabled=false',
      ];
      if (chart === 'note' || chart === 'billing') {
        args.push('--set', 'redis.addr=redis:6379', '--set', 'redis.passwordSecret.name=redis');
      }
      const result = spawnSync('helm', args, {
        input: JSON.stringify({ podSecurityContext: { [field]: null } }),
        encoding: 'utf8',
      });
      assert.notEqual(result.status, 0, `${chart} accepted null ${field}`);
      assert.match(result.stderr, new RegExp(field));
    });
  }
}
