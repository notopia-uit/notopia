// sort-imports-ignore
import 'reflect-metadata';

import datasource from '../src/database/datasource.typeorm';
import { ServerBlockNoteEditor } from '@blocknote/server-util';
import { createBlockNoteSchema } from '@notopia-uit/block-note';

async function run() {
  try {
    console.log('🌱 Initializing DataSource...');
    await datasource.initialize();

    const blockNoteSchema = createBlockNoteSchema('server');
    // TODO: will do with editor later
    ServerBlockNoteEditor.create({ schema: blockNoteSchema });

    console.log('🏃 Running Seeders...');
    // seed

    console.log('✅ Seeding completed!');
    process.exit(0);
  } catch (error) {
    console.error('❌ Seeding failed:', error);
    process.exit(1);
  }
}

void run();
