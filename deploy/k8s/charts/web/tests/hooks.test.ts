import { strict as assert } from 'node:assert';
import { execFileSync } from 'node:child_process';
import { test } from 'node:test';

test('renders the expected hooks without Istio', () => {
  const output = execFileSync(
    'helm',
    ['template', 'test', 'deploy/k8s/charts/web', '--set', 'istio.enabled=false'],
    { encoding: 'utf8' }
  );
  assert.equal((output.match(/^kind: Job$/gm) ?? []).length, 0);
  assert.doesNotMatch(output, /security\.istio\.io/);
});
