import * as fs from 'fs';
import { fileURLToPath } from 'node:url';
import * as path from 'path';

import { NoteSearch } from '@notopia-uit/api-share-gen';
import { Meilisearch } from 'meilisearch';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

const DATA_DIR = path.join(__dirname, 'data');
const MEILI_HOST =
  process.env.NOTOPIA_SEARCH_WORKER_MEILI_HOST || 'http://meilisearch.notopia.localhost';
const MEILI_MASTER_KEY = process.env.NOTOPIA_SEARCH_WORKER_MEILI_API_KEY || 'notopiauit';

async function run() {
  const client = new Meilisearch({
    host: MEILI_HOST,
    apiKey: MEILI_MASTER_KEY,
  });
  const index = client.index('notes');
  const files = fs.readdirSync(DATA_DIR);
  const documents = files.map(
    (file) => JSON.parse(fs.readFileSync(path.join(DATA_DIR, file), 'utf-8')) as NoteSearch
  );

  console.log(`Seeding ${documents.length} documents to Meilisearch...`);
  const taskPromise = index.addDocuments(documents);
  const { taskUid } = await taskPromise;
  console.log(`Documents added with task ID: ${taskUid}`);
  console.log('Waiting for indexing to complete..., you can close this process');
  const task = await taskPromise.waitTask();
  if (task.status === 'succeeded') {
    console.log('Seeding completed successfully!');
  } else {
    console.error(`Seeding failed with status: ${task.status}`, task.error);
    process.exit(1);
  }
}

export { run as SeedMeilisearch };

if (require.main === module) {
  import('./config')
    .then(({ ConfigMeilisearch }) => {
      return ConfigMeilisearch();
    })
    .catch((error) => {
      console.error('Error configuring Meilisearch:', error);
      process.exit(1);
    })
    .then(() => {
      return run();
    })
    .catch((error) => {
      console.error('Error seeding Meilisearch:', error);
      process.exit(2);
    });
}
