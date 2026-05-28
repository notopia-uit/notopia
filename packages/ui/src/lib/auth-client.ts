import { inferAdditionalFields } from 'better-auth/client/plugins';
import { createAuthClient } from 'better-auth/react';

let _client: ReturnType<typeof createAuthClient> | null = null;
let _baseURL = 'http://localhost:3000';

export function configureAuthClient(baseURL: string) {
  _baseURL = baseURL;
  _client = null;
}

export function getAuthClient(): ReturnType<typeof createAuthClient> {
  if (!_client) {
    _client = createAuthClient({
      baseURL: _baseURL,
      plugins: [inferAdditionalFields()],
    });
  }
  return _client;
}
