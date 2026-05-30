import { inferAdditionalFields } from 'better-auth/client/plugins';
import { createAuthClient } from 'better-auth/react';

let _client: ReturnType<typeof createAuthClient> | null = null;
let _baseURL: string | null = null;

export function configureAuthClient(baseURL: string) {
  _baseURL = baseURL;
  _client = null;
}

export function getAuthClient(): ReturnType<typeof createAuthClient> {
  if (!_baseURL) {
    throw new Error(
      'Auth client base URL is not configured. Please call configureAuthClient() first.'
    );
  }
  if (!_client) {
    _client = createAuthClient({
      baseURL: _baseURL,
      plugins: [inferAdditionalFields()],
    });
  }
  return _client;
}
