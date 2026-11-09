import { strict as assert } from 'node:assert';
import { execFileSync, spawnSync } from 'node:child_process';
import { test } from 'node:test';

const chart = 'deploy/k8s/charts/notopia-edge';
const acknowledgements = [
  '--set',
  'security.externalBoundaryAcknowledged=true,security.externalIdentityLayerAcknowledged=true',
];
function render(template: string, ...values: string[]) {
  const yaml = execFileSync(
    'helm',
    [
      'template',
      'test',
      chart,
      ...acknowledgements,
      '--show-only',
      `templates/${template}`,
      ...values,
    ],
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
  const yaml = execFileSync(
    'helm',
    ['template', 'test', chart, ...acknowledgements, '--set', 'istio.enabled=false'],
    {
      encoding: 'utf8',
    }
  );
  const apiVersions: string[] = JSON.parse(
    execFileSync('yq', ['eval-all', '-o=json', '[select(. != null) | .apiVersion]', '-'], {
      input: yaml,
      encoding: 'utf8',
    })
  );
  assert.ok(apiVersions.every((version) => !/^(?:security|telemetry)\.istio\.io\//.test(version)));
});

test('conflicting paths and unsupported wildcards are rejected', () => {
  for (const setting of [
    'istio.publicRules[0].paths[0]=/billing/*',
    'istio.publicRules[0].paths[0]=/*/health/*',
    'istio.groupRules[0].paths[0]=/billing/admin/*,istio.groupRules[0].groups[0]=other',
    'istio.adminGroup=*',
  ]) {
    const result = spawnSync(
      'helm',
      ['template', 'test', chart, ...acknowledgements, '--set', setting],
      {
        encoding: 'utf8',
      }
    );
    assert.notEqual(result.status, 0, setting);
  }
});

test('edge ingress defaults to deny and only permits the configured trusted front gateway', () => {
  const deny = render('networkpolicy.yaml');
  assert.deepEqual(deny.spec.ingress, []);
  const allow = render(
    'networkpolicy.yaml',
    '--set',
    'frontRoute.trustedNamespace=front,frontRoute.trustedGateway=front-gateway,frontRoute.parentRefs[0].name=front-gateway,frontRoute.parentRefs[0].namespace=front'
  );
  assert.deepEqual(allow.spec.ingress[0].from[0], {
    namespaceSelector: { matchLabels: { 'kubernetes.io/metadata.name': 'front' } },
    podSelector: { matchLabels: { 'gateway.networking.k8s.io/gateway-name': 'front-gateway' } },
  });
  const off = render('networkpolicy.yaml', '--set', 'istio.enabled=false');
  assert.deepEqual(off.spec.ingress, []);
});

test('front backend uses the configured listener port', () => {
  assert.equal(
    render('front-httproute.yaml', '--set', 'gateway.listenerPort=8080').spec.rules[0]
      .backendRefs[0].port,
    8080
  );
});

test('rejects unsafe configuration', () => {
  const invalid = [
    ['--set', 'gateway.enabled=false'],
    ['--set', 'frontRoute.trustedNamespace=front'],
    ['--set', 'frontRoute.trustedNamespace=front,frontRoute.trustedGateway=front'],
    [
      '--set',
      'frontRoute.enabled=false,frontRoute.trustedNamespace=front,frontRoute.trustedGateway=front',
    ],
    ...['gateway.enabled', 'frontRoute.enabled', 'istio.enabled', 'istio.telemetry.enabled'].map(
      (s) => ['--set-string', `${s}=false`]
    ),
    ['--set', 'istio.groupRules[0].paths[0]=/secure/%61dmin,istio.groupRules[0].groups[0]=admin'],
    ['--set', 'istio.groupRules[0].paths[0]=/secure/../admin,istio.groupRules[0].groups[0]=admin'],
    ['--set', 'istio.groupRules[0].paths[0]=/secure//admin,istio.groupRules[0].groups[0]=admin'],
    ['--set', 'istio.groupRules[0].paths[0]=/secure/./admin,istio.groupRules[0].groups[0]=admin'],
    ['--set', 'istio.publicRules[0].paths[0]=/public/%2fadmin'],
    ['--set', 'istio.publicRules[0].paths[0]=/x,istio.publicRules[0].method[0]=GET'],
    ['--set', 'istio.publicRules[0].paths[0]=/x,istio.publicRules[0].methods[0]=TYPO'],
    ['--set', 'istio.publicRules[0].methods[0]=GET'],
    ['--set', 'istio.groupRules[0].paths[0]=/x'],
    [
      '--set',
      'istio.groupRules[0].paths[0]=/x,istio.groupRules[0].groups[0]=admin,istio.groupRules[0].group=admin',
    ],
    ['--set', 'istio.jwt.claimHeaders.X-Forwarded-Other=sub'],
    ['--set', 'istio.jwt.claimHeaders.X-Forwarded-Groups=groups'],
  ];
  for (const args of invalid) {
    const result = spawnSync('helm', ['template', 'test', chart, ...acknowledgements, ...args], {
      encoding: 'utf8',
    });
    assert.notEqual(result.status, 0, args.join(' '));
  }
});

test('valid custom group and public rules preserve the authenticated exclusion', () => {
  const rules = render(
    'istio/authorizationpolicy.yaml',
    '--set',
    'istio.groupRules[0].paths[0]=/private/*,istio.groupRules[0].groups[0]=staff,istio.publicRules[0].paths[0]=/open,istio.publicRules[0].methods[0]=GET'
  ).spec.rules;
  assert.ok(
    rules.some(
      (rule: any) =>
        JSON.stringify(rule.to?.[0].operation) ===
        JSON.stringify({ paths: ['/open'], methods: ['GET'] })
    )
  );
  assert.ok(rules.some((rule: any) => rule.when?.[0]?.values?.includes('staff')));
  assert.ok(rules.at(-1).to[0].operation.notPaths.includes('/private/*'));
});
