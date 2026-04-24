import { ShareNoteSearch } from '@notopia-uit/api-gen';
import * as fs from 'fs';
import { Meilisearch } from 'meilisearch';
import { fileURLToPath } from 'node:url';
import * as path from 'path';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

const DATA_DIR = path.join(__dirname, 'data');
const MEILI_HOST =
  process.env.NOTOPIA_SEARCH_WORKER_MEILI_HOST ||
  'http://meilisearch.notopia.localhost';
const MEILI_MASTER_KEY =
  process.env.NOTOPIA_SEARCH_WORKER_MEILI_API_KEY || 'notopiauit';

async function run() {
  const client = new Meilisearch({
    host: MEILI_HOST,
    apiKey: MEILI_MASTER_KEY,
  });
  const index = client.index('notes');
  await index.update({ primaryKey: 'id' }).waitTask();

  console.log('Updating Meilisearch settings...');
  await index
    .updateSettings({
      searchableAttributes: ['name', 'plainTextContent', 'tags'],
      filterableAttributes: ['workspaceId', 'tags'],
      sortableAttributes: ['name'],
    })
    .waitTask();

  const files = fs.readdirSync(DATA_DIR);
  const documents = files.map(
    (file) =>
      JSON.parse(
        fs.readFileSync(path.join(DATA_DIR, file), 'utf-8')
      ) as ShareNoteSearch
  );

  console.log(`Seeding ${documents.length} documents to Meilisearch...`);
  const taskPromise = index.addDocuments(documents);
  const { taskUid } = await taskPromise;
  console.log(`Documents added with task ID: ${taskUid}`);
  console.log(
    'Waiting for indexing to complete..., you can close this process'
  );
  const task = await taskPromise.waitTask();
  if (task.status === 'succeeded') {
    console.log('Seeding completed successfully!');
  } else {
    console.error(`Seeding failed with status: ${task.status}`, task.error);
    process.exit(1);
  }
}

run().catch(console.error);
