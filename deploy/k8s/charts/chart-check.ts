import { execSync } from 'node:child_process';
import { cpSync, mkdirSync, mkdtempSync, readFileSync, rmSync } from 'node:fs';
import path from 'node:path';

const [chart, target] = process.argv.slice(2);
const charts = 'deploy/k8s/charts';
const commands = JSON.parse(readFileSync(`${charts}/chart-commands.json`, 'utf8'));
if (!Object.hasOwn(commands, chart) || !Object.hasOwn(commands[chart], target)) {
  throw new Error(`Unknown chart check: ${chart} ${target}`);
}

mkdirSync('tmp/chart-check', { recursive: true });
const workspace = mkdtempSync(path.resolve('tmp/chart-check/run-'));
try {
  for (const name of new Set(['notopia-common-chart', chart])) {
    const source = path.resolve(charts, name);
    cpSync(source, path.join(workspace, charts, name), {
      recursive: true,
      filter: (file) => path.basename(file) !== 'charts',
    });
  }
  cpSync(`${charts}/chart-commands.json`, path.join(workspace, charts, 'chart-commands.json'));
  for (const command of commands[chart][target]) {
    execSync(command, { cwd: workspace, stdio: 'inherit', shell: '/bin/bash' });
  }
} finally {
  rmSync(workspace, { recursive: true, force: true });
}
