import subprocess
import shutil
import tempfile
import unittest
from pathlib import Path

import yaml

ROOT = Path(__file__).parent
SERVICES = ('web', 'note', 'authorization', 'billing', 'document', 'search-worker')


def build(directory):
    result = subprocess.run(['kubectl', 'kustomize', str(ROOT / directory)], check=True, capture_output=True, text=True)
    return list(yaml.safe_load_all(result.stdout))


class ReleasesTest(unittest.TestCase):
    def test_releases_and_hashed_values(self):
        for directory, names in [('edge', ('notopia-edge',)), ('apps', SERVICES)]:
            documents = build(directory)
            releases = {d['metadata']['name']: d for d in documents if d['kind'] == 'HelmRelease'}
            configmaps = {d['metadata']['name'] for d in documents if d['kind'] == 'ConfigMap'}
            self.assertEqual(set(releases), set(names))
            for name, release in releases.items():
                spec = release['spec']
                self.assertTrue(spec['suspend'])
                self.assertEqual(spec['interval'], '12h')
                self.assertEqual(spec['timeout'], '10m')
                self.assertEqual(spec['install']['remediation']['retries'], 0)
                self.assertEqual(spec['upgrade']['remediation']['strategy'], 'rollback')
                self.assertEqual(spec['chart']['spec']['chart'], './deploy/k8s/charts/' + name)
                self.assertEqual(spec['chart']['spec']['sourceRef'], {'kind': 'GitRepository', 'name': 'notopia', 'namespace': 'flux-system'})
                ref = spec['valuesFrom'][0]['name']
                self.assertIn(ref, configmaps)
                self.assertRegex(ref, '^' + name + '-helm-values-[a-z0-9]{10}$')

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
