import { TerminusModule } from '@nestjs/terminus';
import { Test } from '@nestjs/testing';
import { DataSource } from 'typeorm';
import { expect, it, vi } from 'vitest';

import { authenticatedApiModule } from './authenticated-api.module';
import { SCALING_CONFIG } from './config';
import { HEALTH_REDIS, HealthController } from './health/health.controller';

it('protects every generated API route while health remains public', async () => {
  const attachment = vi.fn((_id: string, _filename: string, request: { user: { id: string } }) => ({
    userId: request.user.id,
  }));
  class DocumentImplementation {
    commitDocument() {
      return {};
    }
    getDocumentAttachmentUploadUrl = attachment;
  }
  class RevisionImplementation {
    deleteRevision() {}
    getRevisionWithContent() {
      return {};
    }
    getRevisions() {
      return {};
    }
    renameRevision() {}
  }
  const module = await Test.createTestingModule({
    imports: [
      authenticatedApiModule({
        apiImplementations: {
          documentApi: DocumentImplementation as never,
          revisionApi: RevisionImplementation as never,
        },
      }),
      TerminusModule,
    ],
    controllers: [HealthController],
    providers: [
      {
        provide: DataSource,
        useValue: {
          query: () => Promise.resolve([]),
          showMigrations: () => Promise.resolve(false),
        },
      },
      { provide: SCALING_CONFIG, useValue: { enabled: false } },
      { provide: HEALTH_REDIS, useValue: { ping: () => Promise.resolve('PONG') } },
    ],
  }).compile();
  const app = module.createNestApplication();
  await app.listen(0, '127.0.0.1');
  try {
    const url = await app.getUrl();
    const routes = [
      ['POST', '/document/documents/one/commit', 201],
      ['GET', '/document/documents/one/attachment-url?filename=a', 200],
      ['DELETE', '/document/revisions/one', 200],
      ['GET', '/document/revisions/one', 200],
      ['GET', '/document/revisions?documentId=one', 200],
      ['POST', '/document/revisions/one/rename', 201],
    ];
    for (const [method, path, expected] of routes as [string, string, number][]) {
      expect((await fetch(url + path, { method })).status, path).toBe(401);
      expect(
        (await fetch(url + path, { method, headers: { 'x-forwarded-id': 'user-1' } })).status,
        path
      ).toBe(expected);
    }
    const response = await fetch(`${url}/document/documents/one/attachment-url?filename=a`, {
      headers: { 'x-forwarded-id': 'user-1' },
    });
    expect(await response.json()).toEqual({ userId: 'user-1' });
    expect((await fetch(`${url}/document/health/live`)).status).toBe(200);
    expect((await fetch(`${url}/document/health/ready`)).status).toBe(200);
    expect((await fetch(`${url}/document/health/startup`)).status).toBe(200);
  } finally {
    await app.close();
  }
});
