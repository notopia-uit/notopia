import assert from 'node:assert/strict';
import { get } from 'node:http';
import { test } from 'node:test';

const base = process.env.GATEWAY_TEST_URL ?? 'http://api.notopia.localhost';

function status(path) {
  return new Promise((resolve, reject) => {
    get(new URL(path, base), { headers: { host: 'api.notopia.localhost' } }, (response) => {
      response.resume();
      resolve(response.statusCode);
    }).on('error', reject);
  });
}

test('document health routes are public while other document routes require authentication', async () => {
  for (const path of [
    '/document/health/live',
    '/document/health/ready',
    '/document/health/startup',
  ]) {
    assert.equal(await status(path), 200, path);
  }

  for (const path of [
    '/document/documents',
    '/document/healthcheck',
    '/document/health-private',
    '/document/health/live/private',
  ]) {
    assert.equal(await status(path), 401, path);
  }
});

test('document websocket route remains public', async () => {
  assert.equal(await status('/document/ws/document'), 200);
});
