// oxfmt-ignore
import 'reflect-metadata';

import { type MySchema } from '@blocknote/core';
import { ServerBlockNoteEditor } from '@blocknote/server-util';
import { createSchema } from '@notopia-uit/lib-server/block-note';

import datasource from './datasource';
import DocumentSeeder from './seeds/documents';
import RevisionSeeder from './seeds/revisions';

async function run() {
  try {
    console.log('🌱 Initializing DataSource...');
    await datasource.initialize();

    const blockNoteSchema: MySchema = createSchema();
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

if (import.meta.main) {
  run().catch((error) => {
    console.error('Error during seeding:', error);
    process.exit(1);
  });
}
