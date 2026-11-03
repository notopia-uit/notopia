import { NoteSearch } from '@notopia-uit/api-share-gen';
import { Meilisearch, MeilisearchApiError, type Settings } from 'meilisearch';
import pino from 'pino';

const logger = pino();

async function run() {
  const host = process.env.NOTOPIA_SEARCH_WORKER_MEILI_HOST;
  const apiKey = process.env.NOTOPIA_SEARCH_WORKER_MEILI_API_KEY;
  const searchApiKey = process.env.NOTOPIA_SEARCH_WORKER_MEILI_NOTE_SEARCH_KEY_UID;
  if (!host || !apiKey || !searchApiKey) {
    throw new Error(
      'NOTOPIA_SEARCH_WORKER_MEILI_HOST, NOTOPIA_SEARCH_WORKER_MEILI_API_KEY and NOTOPIA_SEARCH_WORKER_MEILI_NOTE_SEARCH_KEY_UID are required'
    );
  }
  const client = new Meilisearch({
    host,
    apiKey,
  });
  const index = client.index('notes');
  logger.info('Ensuring index exists and has correct primary key...');
  try {
    const indexInfo = await index.getRawInfo();
    if (indexInfo.primaryKey !== 'id') {
      logger.info(`Index primary key is '${indexInfo.primaryKey}', updating to 'id'...`);
      await index.update({ primaryKey: 'id' }).waitTask();
    } else {
      logger.info('Index primary key is already set to "id", skipping update.');
    }
  } catch (error) {
    if (error instanceof MeilisearchApiError && error.response.status === 404) {
      logger.info('Index not found, creating it with primary key "id"...');
      await index.update({ primaryKey: 'id' }).waitTask();
      logger.info('Index created successfully.');
    } else {
      throw error;
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

  logger.info('Get current index settings...');
  try {
    const existedSettings = await index.getSettings();
    logger.info('Checking if existing settings match desired settings...');
    const settingsMatch = JSON.stringify(existedSettings) === JSON.stringify(settings);
    if (settingsMatch) {
      logger.info('Existing settings match desired settings, skipping update.');
    } else {
      logger.info('Existing settings do not match desired settings, updating...');
      await index.updateSettings(settings).waitTask();
      logger.info('Index settings updated successfully.');
    }
  } catch (error) {
    if (error instanceof MeilisearchApiError && error.response.status === 404) {
      logger.info('Index not found, creating it with settings...');
      await index.updateSettings(settings).waitTask();
      logger.info('Index created successfully.');
    } else {
      throw error;
    }
  }

  logger.info('Check if default search API key exists...');
  try {
    await client.getKey(searchApiKey);
    logger.info('Default search API key already exists, skipping creation.');
  } catch (error) {
    if (error instanceof MeilisearchApiError && error.response.status === 404) {
      logger.info('Default search API key not found, creating it...');
      await client.createKey({
        uid: searchApiKey,
        actions: ['search'],
        indexes: ['notes'],
        expiresAt: null,
        name: 'Default Note Index Search Key',
      });
      logger.info('Default search API key created successfully');
    } else {
      throw error;
    }
  }
}

export { run as ConfigMeilisearch };
