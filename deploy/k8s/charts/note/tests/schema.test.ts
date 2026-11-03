import { strict as assert } from 'node:assert';
import { spawnSync } from 'node:child_process';
import { basename, dirname } from 'node:path';
import { test } from 'node:test';
import { fileURLToPath } from 'node:url';

const chart = dirname(dirname(fileURLToPath(import.meta.url)));
const name = basename(chart);
const base = ['template', 'test', chart, '--set', 'istio.enabled=false'];
if (name !== 'authorization')
  base.push('--set', 'redis.addr=redis:6379', '--set', 'redis.passwordSecret.name=redis');

test('accepts CPU limits and extended resource quantities', () => {
  const result = spawnSync(
    'helm',
    [
      ...base,
      '--set',
      'resources.limits.cpu=250m',
      '--set',
      'resources.limits.example\\.com/gpu=1',
      '--set',
      'resources.requests.example\\.com/gpu=1',
    ],
    { encoding: 'utf8' }
  );
  assert.equal(result.status, 0, result.stderr);
  assert.match(result.stdout, /example\.com\/gpu/);
});

for (const [key, value] of [
  ['replicaCount', 'oops'],
  ['autoscaling.enabled', 'nope'],
  ['autoscaling.minReplicas', '0'],
  ['route.hostnames', 'oops'],
  ['image.tag', '42'],
  ['hooks.seed.activeDeadlineSeconds', '-1'],
] as const) {
  test(`rejects invalid ${key}`, () => {
    const result = spawnSync('helm', [...base, '--set', `${key}=${value}`], { encoding: 'utf8' });
    assert.notEqual(result.status, 0, result.stdout);
    assert.match(result.stderr, /values don't meet the specifications/);
  });
}

test('accepts percentage disruption budgets', () => {
  const result = spawnSync(
    'helm',
    [
      ...base,
      '--set',
      'podDisruptionBudget.enabled=true',
      '--set',
      'podDisruptionBudget.minAvailable=50%',
    ],
    { encoding: 'utf8' }
  );
  assert.equal(result.status, 0, result.stderr);
});
