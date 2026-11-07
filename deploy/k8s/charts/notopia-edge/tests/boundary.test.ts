import { strict as assert } from 'node:assert';
import { spawnSync } from 'node:child_process';
import { test } from 'node:test';

const chart = 'deploy/k8s/charts/notopia-edge';
function helm(...args: string[]) {
  return spawnSync('helm', ['template', 'test', chart, ...args], { encoding: 'utf8' });
}

test('default install and Istio-off install refuse an unacknowledged external boundary', () => {
  for (const args of [[], ['--set', 'istio.enabled=false']]) {
    const result = helm(...args);
    assert.notEqual(result.status, 0);
    assert.match(result.stderr, /security\.externalBoundaryAcknowledged/);
  }
});

test('Istio-off requires a separate acknowledgement of the external identity layer', () => {
  const result = helm('--set', 'security.externalBoundaryAcknowledged=true,istio.enabled=false');
  assert.notEqual(result.status, 0);
  assert.match(result.stderr, /security\.externalIdentityLayerAcknowledged/);
});

test('explicit boolean acknowledgements permit rendering in both modes, not string values', () => {
  assert.equal(helm('--set', 'security.externalBoundaryAcknowledged=true').status, 0);
  assert.equal(
    helm(
      '--set',
      'security.externalBoundaryAcknowledged=true,security.externalIdentityLayerAcknowledged=true,istio.enabled=false'
    ).status,
    0
  );
  assert.notEqual(helm('--set-string', 'security.externalBoundaryAcknowledged=true').status, 0);
  assert.notEqual(
    helm(
      '--set',
      'security.externalBoundaryAcknowledged=true,istio.enabled=false',
      '--set-string',
      'security.externalIdentityLayerAcknowledged=true'
    ).status,
    0
  );
  assert.notEqual(
    helm('--skip-schema-validation', '--set-string', 'security.externalBoundaryAcknowledged=true')
      .status,
    0
  );
  assert.notEqual(
    helm(
      '--skip-schema-validation',
      '--set',
      'security.externalBoundaryAcknowledged=true,istio.enabled=false',
      '--set-string',
      'security.externalIdentityLayerAcknowledged=true'
    ).status,
    0
  );
});
