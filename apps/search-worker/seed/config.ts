import { NoteSearch } from '@notopia-uit/api-share-gen';
import { Meilisearch, MeilisearchApiError, type Settings } from 'meilisearch';

const MEILI_HOST =
  process.env.NOTOPIA_SEARCH_WORKER_MEILI_HOST || 'http://meilisearch.notopia.localhost';
const MEILI_MASTER_KEY = process.env.NOTOPIA_SEARCH_WORKER_MEILI_API_KEY || 'notopiauit';

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
      console.log(`Index primary key is '${indexInfo.primaryKey}', updating to 'id'...`);
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
    searchableAttributes: [
      'name',
      'folderName',
      'plainTextContent',
      'tags',
    ] satisfies (keyof NoteSearch)[],
    filterableAttributes: ['workspaceId', 'tags'] satisfies (keyof NoteSearch)[],
    sortableAttributes: ['name'] satisfies (keyof NoteSearch)[],
  };

  console.log('Get current index settings...');
  try {
    const existedSettings = await index.getSettings();
    console.log('Checking if existing settings match desired settings...');
    const settingsMatch = JSON.stringify(existedSettings) === JSON.stringify(settings);
    if (settingsMatch) {
      console.log('Existing settings match desired settings, skipping update.');
    } else {
      console.log('Existing settings do not match desired settings, updating...');
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
  const searchApiKey = '00000000-0000-4000-8000-000000000001';
  try {
    await client.getKey(searchApiKey);
    console.log('Default search API key already exists, skipping creation.');
  } catch (error) {
    if (error instanceof MeilisearchApiError) {
      console.log('Default search API key not found, creating it...');
      const key = await client.createKey({
        uid: searchApiKey,
        actions: ['search'],
        indexes: ['notes'],
        expiresAt: null,
        name: 'Default Note Index Search Key',
      });
      console.log('Default search API key created successfully');
      console.log(`Key UID: ${key.uid}`);
      console.log(`Key: ${key.key}`);
      console.log(
        `Please save the Key to cmd/note/.env.local as
NOTOPIA_NOTE_MEILISEARCH_API_KEY=${key.key}`
      );
    } else {
      console.error('Error checking/creating default search API key:', error);
      process.exit(1);
    }
  }
}

export { run as ConfigMeilisearch };

if (require.main === module) {
  try {
    void run();
  } catch (error) {
    console.error('Error configuring Meilisearch:', error);
    process.exit(1);
  }
}
