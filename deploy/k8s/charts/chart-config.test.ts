import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import { test } from 'node:test';

for (const chart of ['notopia-common-chart', 'note', 'authorization', 'billing']) {
  test(`${chart} chart checks use isolated workspaces and track shared inputs`, () => {
    const project = JSON.parse(readFileSync(`deploy/k8s/charts/${chart}/project.json`, 'utf8'));
    for (const target of ['lint', 'test']) {
      assert.notEqual(project.targets[target].parallelism, false);
      assert.match(project.targets[target].options.command, /chart-check\.ts/);
      assert.ok(project.targets[target].inputs.includes('chartInputs'));
    }
  });
}
