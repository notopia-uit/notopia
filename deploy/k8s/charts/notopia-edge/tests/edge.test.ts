import { strict as assert } from 'node:assert';
import { execFileSync, spawnSync } from 'node:child_process';
import { test } from 'node:test';

const chart = 'deploy/k8s/charts/notopia-edge';
function render(template: string, ...values: string[]) {
  const yaml = execFileSync(
    'helm',
    ['template', 'test', chart, '--show-only', `templates/${template}`, ...values],
    { encoding: 'utf8' }
  );
  return JSON.parse(execFileSync('yq', ['-o=json'], { input: yaml, encoding: 'utf8' }));
}

test('front route strips all identity headers', () => {
  const remove =
    render('front-httproute.yaml').spec.rules[0].filters[0].requestHeaderModifier.remove;
  assert.deepEqual(remove, [
    'X-Forwarded-ID',
    'X-Forwarded-Email',
    'X-Forwarded-Name',
    'X-Forwarded-Groups',
    'X-Forwarded-Roles',
  ]);
});

test('groups header uses the scalar groups_json claim', () => {
  const headers = render('istio/requestauthentication.yaml').spec.jwtRules[0].outputClaimToHeaders;
  assert.ok(
    headers.some(
      (entry: { header: string; claim: string }) =>
        entry.header === 'X-Forwarded-Groups' && entry.claim === 'groups_json'
    )
  );
});

test('public and protected policies and authenticated fallback are disjoint', () => {
  const rules = render('istio/authorizationpolicy.yaml').spec.rules;
  const paths = rules[0].to[0].operation.paths;
  for (const svc of ['note', 'authorization', 'billing', 'document', 'search-worker'])
    assert.ok(paths.includes(`/${svc}/health/*`));
  assert.ok(paths.includes('/billing/webhooks/stripe'));
  assert.ok(paths.includes('/document/ws/document'));
  assert.deepEqual(rules[1].to[0].operation, { paths: ['/billing/plans'], methods: ['GET'] });
  assert.deepEqual(rules[2].when[0], { key: 'request.auth.claims[groups]', values: ['admin'] });
  assert.deepEqual(rules[2].to[0].operation.paths, ['/billing/admin', '/billing/admin/*']);
  assert.deepEqual(rules[3].from[0].source.requestPrincipals, ['*']);
  assert.deepEqual(rules[3].to[0].operation.notPaths, ['/billing/admin', '/billing/admin/*']);
});

test('Istio disabled emits no Istio objects', () => {
  const yaml = execFileSync('helm', ['template', 'test', chart, '--set', 'istio.enabled=false'], {
    encoding: 'utf8',
  });
  assert.doesNotMatch(yaml, /(?:security|telemetry)\.istio\.io/);
});

test('conflicting paths and unsupported wildcards are rejected', () => {
  for (const setting of [
    'istio.publicRules[0].paths[0]=/billing/*',
    'istio.publicRules[0].paths[0]=/*/health/*',
    'istio.groupRules[0].paths[0]=/billing/admin/*,istio.groupRules[0].groups[0]=other',
    'istio.adminGroup=*',
  ]) {
    const result = spawnSync('helm', ['template', 'test', chart, '--set', setting], {
      encoding: 'utf8',
    });
    assert.notEqual(result.status, 0, setting);
  }
});
