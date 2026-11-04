import { strict as assert } from 'node:assert';
import { execFileSync } from 'node:child_process';
import { test } from 'node:test';

type Route = {
  spec: {
    rules: Array<{
      matches: Array<{ path: { type: string; value: string } }>;
      backendRefs?: Array<{ name: string; port: number }>;
    }>;
  };
};

function route(exposeHealth: boolean): Route {
  const yaml = execFileSync(
    'helm',
    [
      'template',
      'test',
      'deploy/k8s/charts/document',
      '--show-only',
      'templates/httproute.yaml',
      '--set',
      'istio.enabled=false',
      '--set',
      `route.exposeHealth=${exposeHealth}`,
    ],
    { encoding: 'utf8' }
  );
  return JSON.parse(execFileSync('yq', ['-o=json'], { input: yaml, encoding: 'utf8' }));
}

function backendForPath(route: Route, path: string) {
  const matches = route.spec.rules.flatMap((rule) =>
    rule.matches
      .filter(
        ({ path: match }) =>
          match.type === 'PathPrefix' &&
          (path === match.value || path.startsWith(`${match.value}/`))
      )
      .map(({ path: match }) => ({ prefix: match.value, backend: rule.backendRefs?.[0] }))
  );
  return matches.sort((a, b) => b.prefix.length - a.prefix.length)[0]?.backend;
}

test('the document prefix still routes health when the explicit health route is disabled', () => {
  const rendered = route(false);
  assert.equal(backendForPath(rendered, '/document/documents/123')?.port, 8082);
  assert.equal(backendForPath(rendered, '/document/ws/document')?.port, 8082);
  assert.equal(backendForPath(rendered, '/document/health/ready')?.port, 8082);
  assert.equal(backendForPath(rendered, '/document/health')?.port, 8082);
  assert.equal(backendForPath(rendered, '/documentary/health'), undefined);
});

test('the explicit health route forwards health and leaves API routing intact', () => {
  const rendered = route(true);
  assert.equal(backendForPath(rendered, '/document/health/ready')?.port, 8082);
  assert.equal(backendForPath(rendered, '/document/documents/123')?.port, 8082);
  assert.equal(backendForPath(rendered, '/document/ws/document')?.port, 8082);
});
