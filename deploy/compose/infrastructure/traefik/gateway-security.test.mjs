import assert from 'node:assert/strict';
import { test } from 'node:test';

// Point these routes at a disposable backend which returns request headers as JSON.
const base = process.env.GATEWAY_TEST_URL ?? 'http://api.notopia.localhost';
const normalToken = process.env.GATEWAY_TEST_USER_TOKEN;
const adminToken = process.env.GATEWAY_TEST_ADMIN_TOKEN;
const expectedGroups = process.env.GATEWAY_TEST_USER_GROUPS;

async function request(path, token) {
  const headers = {
    'X-Forwarded-ID': 'forged',
    'X-Forwarded-Groups': '["notopia-admins"]',
    'X-Forwarded-Roles': '["admin"]',
  };
  if (token) headers.Authorization = `Bearer ${token}`;
  return fetch(`${base}${path}`, { headers });
}

test('gateway rejects forged identity without a token', async () => {
  for (const path of ['/billing/test-headers', '/billing/admin/test-headers']) {
    assert.equal((await request(path)).status, 401);
  }
});

test('only admins can access admin routes', async () => {
  assert.ok(normalToken && adminToken, 'set GATEWAY_TEST_USER_TOKEN and GATEWAY_TEST_ADMIN_TOKEN');
  assert.equal((await request('/billing/admin/test-headers', normalToken)).status, 403);
  assert.equal((await request('/billing/admin/test-headers', adminToken)).status, 200);
});

test('gateway replaces forged groups with the authenticated user groups', async () => {
  assert.ok(normalToken, 'set GATEWAY_TEST_USER_TOKEN');
  assert.ok(expectedGroups, 'set GATEWAY_TEST_USER_GROUPS to a JSON array');
  const response = await request('/billing/test-headers', normalToken);
  assert.equal(response.status, 200);
  const headers = await response.json();
  assert.deepEqual(JSON.parse(headers['x-forwarded-groups']), JSON.parse(expectedGroups));
  assert.notEqual(headers['x-forwarded-id'], 'forged');
  assert.equal(headers['x-forwarded-roles'], undefined);
});

test('gateway removes forged groups when the token has no group claim', async () => {
  assert.ok(process.env.GATEWAY_TEST_NO_GROUPS_TOKEN, 'set GATEWAY_TEST_NO_GROUPS_TOKEN');
  const response = await request('/billing/test-headers', process.env.GATEWAY_TEST_NO_GROUPS_TOKEN);
  assert.equal(response.status, 200);
  const headers = await response.json();
  assert.equal(headers['x-forwarded-groups'], undefined);
});
