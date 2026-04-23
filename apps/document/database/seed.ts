// sort-imports-ignore
import 'reflect-metadata';

import datasource from './datasource';
import { ServerBlockNoteEditor } from '@blocknote/server-util';
import { type MySchema } from '@blocknote/core';
import { createServerBlockNoteSchema } from '@notopia-uit/lib/server';
import DocumentSeeder from './seeds/documents';
import RevisionSeeder from './seeds/revisions';

async function run() {
  try {
    console.log('🌱 Initializing DataSource...');
    await datasource.initialize();

    const blockNoteSchema: MySchema = createServerBlockNoteSchema();
    // TODO: will do with editor later
    const editor = ServerBlockNoteEditor.create({
      schema: blockNoteSchema,
    });

    console.log('🏃 Running Seeders...');
    const documentSeeder = new DocumentSeeder(editor);
    const revisionSeeder = new RevisionSeeder(editor);

    await documentSeeder.run(datasource);
    await revisionSeeder.run(datasource);

    console.log('✅ Seeding completed!');
    process.exit(0);
  } catch (error) {
    console.error('❌ Seeding failed:', error);
    process.exit(1);
  }
}

void run();
