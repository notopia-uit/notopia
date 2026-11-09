import { strict as assert } from 'node:assert';
import { execFileSync } from 'node:child_process';
import { test } from 'node:test';

test('deployment renders with the Kubernetes API version', () => {
  const yaml = execFileSync(
    'helm',
    [
      'template',
      'test',
      'deploy/k8s/charts/search-worker',
      '--show-only',
      'templates/deployment.yaml',
      '--set',
      'istio.enabled=false',
    ],
    { encoding: 'utf8' }
  );
  const deployment = JSON.parse(execFileSync('yq', ['-o=json'], { input: yaml, encoding: 'utf8' }));
  assert.equal(deployment.apiVersion, 'apps/v1');
  assert.equal(deployment.kind, 'Deployment');
});
