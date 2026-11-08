import subprocess
import shutil
import tempfile
import unittest
from pathlib import Path

import yaml

ROOT = Path(__file__).parent
SERVICES = ('web', 'note', 'authorization', 'billing', 'document', 'search-worker')
BROKER = 'redpanda.redpanda.svc.cluster.local:9093'


def build(directory):
    result = subprocess.run(['kubectl', 'kustomize', str(ROOT / directory)], check=True, capture_output=True, text=True)
    return list(yaml.safe_load_all(result.stdout))


class ReleasesTest(unittest.TestCase):
    def test_broker_values(self):
        for name in ('note', 'authorization', 'billing', 'document', 'search-worker'):
            values = yaml.safe_load((ROOT / 'apps' / name / 'values.yaml').read_text())
            key = 'NOTOPIA_' + name.upper().replace('-', '_') + '_KAFKA_BROKERS'
            self.assertEqual(values['env'][key], BROKER, name)

    def test_exact_principals_and_rendered_ports(self):
        edge = 'cluster.local/ns/notopia/sa/notopia-edge-istio'
        principal = lambda name: 'cluster.local/ns/notopia/sa/' + name
        expected = {
            'web': ({'cluster.local/ns/istio-system/sa/default-istio'}, {'3000'}),
            'note': ({edge, principal('document')}, {'8081', '18081', '28081'}),
            'authorization': ({edge, principal('note'), principal('document'), principal('billing')}, {'18089', '28089'}),
            'billing': ({edge, principal('note'), principal('document')}, {'8084', '18084', '28084'}),
            'document': ({edge}, {'8082'}),
            'search-worker': ({edge}, {'28083'}),
        }
        for name, (principals, ports) in expected.items():
            result = subprocess.run(['helm', 'template', name, str(ROOT.parents[1] / 'charts' / name), '-n', 'notopia', '-f', str(ROOT / 'apps' / name / 'values.yaml')], check=True, capture_output=True, text=True)
            resources = [d for d in yaml.safe_load_all(result.stdout) if d]
            policy = next(d for d in resources if d['kind'] == 'AuthorizationPolicy')
            rule = policy['spec']['rules'][0]
            self.assertEqual(set(rule['from'][0]['source']['principals']), principals, name)
            self.assertEqual(set(rule['to'][0]['operation']['ports']), ports, name)
            service = next(d for d in resources if d['kind'] == 'Service')
            self.assertEqual({str(p['port']) for p in service['spec']['ports']}, ports, name)

    def test_rendered_public_rules_and_header_removal(self):
        chart = ROOT.parents[1] / 'charts/notopia-edge'
        result = subprocess.run(['helm', 'template', 'notopia-edge', str(chart), '-n', 'notopia', '-f', str(ROOT / 'edge/values.yaml'), '--set', 'security.externalBoundaryAcknowledged=true'], check=True, capture_output=True, text=True)
        resources = [d for d in yaml.safe_load_all(result.stdout) if d]
        policy = next(d for d in resources if d['kind'] == 'AuthorizationPolicy')
        public = set()
        for rule in policy['spec']['rules']:
            if 'from' not in rule:
                operation = rule['to'][0]['operation']
                for path in operation['paths']:
                    public.add((path, tuple(operation.get('methods', []))))
        self.assertEqual(public, {(p, ()) for p in ['/note/health/*', '/authorization/health/*', '/billing/health/*', '/document/health/*', '/search-worker/health/*', '/billing/webhooks/stripe', '/document/ws/document']} | {('/billing/plans', ('GET',))})
        route = next(d for d in resources if d['kind'] == 'HTTPRoute')
        filters = route['spec']['rules'][0]['filters']
        modifier = next(f['requestHeaderModifier'] for f in filters if f['type'] == 'RequestHeaderModifier')
        self.assertEqual(set(modifier['remove']), {'X-Forwarded-ID', 'X-Forwarded-Email', 'X-Forwarded-Name', 'X-Forwarded-Groups', 'X-Forwarded-Roles'})

    def test_releases_and_hashed_values(self):
        for directory, names in [('edge', ('notopia-edge',)), ('apps', SERVICES)]:
            documents = build(directory)
            releases = {d['metadata']['name']: d for d in documents if d['kind'] == 'HelmRelease'}
            configmaps = {d['metadata']['name']: d for d in documents if d['kind'] == 'ConfigMap'}
            self.assertEqual(set(releases), set(names))
            for name, release in releases.items():
                spec = release['spec']
                self.assertTrue(spec['suspend'])
                self.assertEqual(spec['interval'], '12h')
                self.assertEqual(spec['timeout'], '10m')
                self.assertEqual(spec['install']['remediation']['retries'], 0)
                self.assertEqual(spec['upgrade']['remediation']['strategy'], 'rollback')
                self.assertEqual(spec['upgrade']['remediation']['retries'], 0)
                self.assertEqual(spec['upgrade']['remediation']['remediateLastFailure'], name != 'authorization')
                self.assertEqual(spec['chart']['spec']['chart'], './deploy/k8s/charts/' + name)
                self.assertEqual(spec['chart']['spec']['sourceRef'], {'kind': 'GitRepository', 'name': 'notopia', 'namespace': 'flux-system'})
                ref = spec['valuesFrom'][0]['name']
                self.assertIn(ref, configmaps)
                self.assertRegex(ref, '^' + name + '-helm-values-[a-z0-9]{10}$')
                source = ROOT / directory / ('values.yaml' if directory == 'edge' else name + '/values.yaml')
                self.assertEqual(yaml.safe_load(configmaps[ref]['data']['values.yaml']), yaml.safe_load(source.read_text()))

    def test_wiring(self):
        values = {name: yaml.safe_load((ROOT / 'apps' / name / 'values.yaml').read_text()) for name in SERVICES}
        for name, value in values.items():
            self.assertIn('tag', value['image'])
            self.assertEqual(value['otel']['endpoint'], 'http://alloy-alloy.monitoring.svc:4317')
        self.assertTrue(values['web']['env']['DOCUMENT_WS_URL'].startswith('wss://'))
        for name in ('note', 'billing', 'authorization'):
            self.assertEqual(values[name]['database']['urlSecret']['name'], 'cnpg-' + name + '-app')
        for name in ('note', 'billing'):
            self.assertEqual(values[name]['redis']['addr'], 'valkey-' + name + ':6379')
            self.assertEqual(values[name]['redis']['passwordSecret']['name'], 'valkey-' + name)
        self.assertEqual(values['search-worker']['envFrom'], [{'secretRef': {'name': 'search-worker'}}])
        self.assertEqual(values['note']['env']['NOTOPIA_NOTE_MEILISEARCH_NOTE_SEARCH_KEY_UID'], values['search-worker']['env']['NOTOPIA_SEARCH_WORKER_MEILI_NOTE_SEARCH_KEY_UID'])
        self.assertEqual({e['valueFrom']['secretKeyRef']['key'] for e in values['document']['extraEnv']}, {'host', 'port', 'username', 'password', 'dbname'})
        expected_addresses = {
            'note': {
                'NOTOPIA_NOTE_SERVICES_AUTHORIZATION_URL': 'authorization:18089',
                'NOTOPIA_NOTE_SERVICES_AUTHORIZATION_LIVE_URL': 'http://authorization:28089/authorization/health/live',
                'NOTOPIA_NOTE_SERVICES_BILLING_URL': 'billing:18084',
                'NOTOPIA_NOTE_SERVICES_BILLING_LIVE_URL': 'http://billing:28084/billing/health/live',
                'NOTOPIA_NOTE_MEILISEARCH_HOST': 'http://meilisearch:7700',
            },
            'billing': {
                'NOTOPIA_BILLING_SERVICES_AUTHORIZATION_URL': 'authorization:18089',
                'NOTOPIA_BILLING_SERVICES_AUTHORIZATION_LIVE_URL': 'http://authorization:28089/authorization/health/live',
            },
            'document': {
                'NOTOPIA_DOCUMENT_SERVICES_NOTE_GRPC_URL': 'note:18081',
                'NOTOPIA_DOCUMENT_SERVICES_AUTHORIZATION_GRPC_URL': 'authorization:18089',
                'NOTOPIA_DOCUMENT_SERVICES_BILLING_GRPC_URL': 'billing:18084',
            },
            'search-worker': {'NOTOPIA_SEARCH_WORKER_MEILI_HOST': 'http://meilisearch:7700'},
        }
        for name, addresses in expected_addresses.items():
            for key, address in addresses.items():
                self.assertEqual(values[name]['env'][key], address)
        for name, callers in {'note': ['document'], 'billing': ['note', 'document'], 'authorization': ['note', 'document', 'billing']}.items():
            for caller in callers:
                self.assertIn('cluster.local/ns/notopia/sa/' + caller, values[name]['istio']['allowedPrincipals'])
        edge = yaml.safe_load((ROOT / 'edge/values.yaml').read_text())
        self.assertFalse(edge['security']['externalBoundaryAcknowledged'])
        self.assertEqual(edge['istio']['adminGroup'], 'notopia-admins')
        for rule in edge['istio']['publicRules']:
            for path in rule['paths']:
                self.assertNotIn('*', path[:-1])
                self.assertFalse(path.startswith('/billing/admin'))

    def test_rendered_edge_excludes_admin_from_general_jwt(self):
        chart = ROOT.parents[1] / 'charts/notopia-edge'
        result = subprocess.run(['helm', 'template', 'notopia-edge', str(chart), '-n', 'notopia', '-f', str(ROOT / 'edge/values.yaml'), '--set', 'security.externalBoundaryAcknowledged=true'], check=True, capture_output=True, text=True)
        policy = next(d for d in yaml.safe_load_all(result.stdout) if d and d['kind'] == 'AuthorizationPolicy')
        rules = policy['spec']['rules']
        general = next(r for r in rules if r.get('to', [{}])[0].get('operation', {}).get('notPaths'))
        self.assertEqual(set(general['to'][0]['operation']['notPaths']), {'/billing/admin', '/billing/admin/*'})
        admin = next(r for r in rules if r.get('when'))
        self.assertEqual(admin['when'][0]['values'], ['notopia-admins'])

    def test_values_changes_rename_configmaps(self):
        with tempfile.TemporaryDirectory(dir=ROOT) as temp:
            for directory, name, values_path in [('edge', 'notopia-edge', 'values.yaml'), ('apps', 'note', 'note/values.yaml')]:
                target = Path(temp) / directory
                shutil.copytree(ROOT / directory, target)
                def reference():
                    result = subprocess.run(['kubectl', 'kustomize', str(target)], check=True, capture_output=True, text=True)
                    return next(d['spec']['valuesFrom'][0]['name'] for d in yaml.safe_load_all(result.stdout) if d['kind'] == 'HelmRelease' and d['metadata']['name'] == name)
                before = reference()
                with (target / values_path).open('a') as f:
                    f.write('\n# changed values\n')
                self.assertNotEqual(before, reference())


if __name__ == '__main__':
    unittest.main()
