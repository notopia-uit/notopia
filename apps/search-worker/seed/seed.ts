import { ShareNoteSearch } from '@notopia-uit/api-gen';
import * as fs from 'fs';
import { Meilisearch, MeilisearchApiError, type Settings } from 'meilisearch';
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
  console.log('Ensuring index exists and has correct primary key...');
  try {
    const indexInfo = await index.getRawInfo();
    if (indexInfo.primaryKey !== 'id') {
      console.log(
        `Index primary key is '${indexInfo.primaryKey}', updating to 'id'...`
      );
      await index.update({ primaryKey: 'id' }).waitTask();
    } else {
      console.log('Index primary key is already set to "id", skipping update.');
    }
  } catch (error) {
    if (error instanceof MeilisearchApiError) {
      console.log('Index not found, creating it with primary key "id"...');
      await index.update({ primaryKey: 'id' }).waitTask();
      console.log('Index created successfully.');
    } else {
      console.error('Error checking/creating index:', error);
      process.exit(1);
    }
  }

  const settings: Settings = {
    searchableAttributes: ['name', 'plainTextContent', 'tags'],
    filterableAttributes: ['workspaceId', 'tags'],
    sortableAttributes: ['name'],
  };

  console.log('Get current index settings...');
  try {
    const existedSettings = await index.getSettings();
    console.log('Checking if existing settings match desired settings...');
    const settingsMatch =
      JSON.stringify(existedSettings) === JSON.stringify(settings);
    if (settingsMatch) {
      console.log('Existing settings match desired settings, skipping update.');
    } else {
      console.log(
        'Existing settings do not match desired settings, updating...'
      );
      await index.updateSettings(settings).waitTask();
      console.log('Index settings updated successfully.');
    }
  } catch (error) {
    if (error instanceof MeilisearchApiError) {
      console.log('Index not found, creating it with settings...');
      await index.updateSettings(settings).waitTask();
      console.log('Index created successfully.');
    } else {
      console.error('Error checking/creating index:', error);
      process.exit(1);
    }
  }

  console.log('Check if default search API key exists...');
  try {
    await client.getKey('00000000-0000-4000-0000-000000000000');
    console.log('Default search API key already exists, skipping creation.');
  } catch (error) {
    if (error instanceof MeilisearchApiError) {
      console.log('Default search API key not found, creating it...');
      await client.createKey({
        uid: '00000000-0000-4000-0000-000000000000',
        actions: ['search'],
        indexes: ['notes'],
        expiresAt: null,
        name: 'Default Search Key',
      });
      console.log('Default search API key created successfully.');
    } else {
      console.error('Error checking/creating default search API key:', error);
      process.exit(1);
    }
  }

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
