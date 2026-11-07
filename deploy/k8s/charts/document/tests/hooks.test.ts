import { strict as assert } from 'node:assert';
import { execFileSync } from 'node:child_process';
import { test } from 'node:test';

test('renders the expected hooks without Istio', () => {
  const output = execFileSync(
    'helm',
    ['template', 'test', 'deploy/k8s/charts/document', '--set', 'istio.enabled=false'],
    { encoding: 'utf8' }
  );
  assert.equal((output.match(/^kind: Job$/gm) ?? []).length, 1);
  assert.doesNotMatch(output, /security\.istio\.io/);
  assert.match(output, /helm.sh\/hook: "pre-install,pre-upgrade"/);
  assert.match(output, /- migrate.cjs/);
});

test('migration and deployment receive the same Redis environment', () => {
  const output = execFileSync(
    'helm',
    [
      'template',
      'test',
      'deploy/k8s/charts/document',
      '--set',
      'istio.enabled=false',
      '--set',
      'scaling.enabled=true',
      '--set',
      'scaling.redisUrlSecret.name=redis',
    ],
    { encoding: 'utf8' }
  );
  assert.equal((output.match(/name: NOTOPIA_DOCUMENT_REDIS_URL/g) ?? []).length, 2);
  assert.equal((output.match(/name: NOTOPIA_DOCUMENT_SCALING_ENABLED/g) ?? []).length, 2);
});
