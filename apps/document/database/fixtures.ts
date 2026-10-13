// oxfmt-ignore
import 'reflect-metadata';

import { type MySchema } from '@blocknote/core';
import { ServerBlockNoteEditor } from '@blocknote/server-util';
import { createSchema } from '@notopia-uit/lib-server/block-note';

import datasource from './datasource';
import DocumentFixtures from './fixtures/documents';
import RevisionFixtures from './fixtures/revisions';

async function run() {
  try {
    console.log('🌱 Initializing DataSource...');
    await datasource.initialize();

    const blockNoteSchema: MySchema = createSchema();
    // TODO: will do with editor later
    const editor = ServerBlockNoteEditor.create({
      schema: blockNoteSchema,
    });

    console.log('🏃 Running fixtures...');
    const documentFixtures = new DocumentFixtures(editor);
    const revisionFixtures = new RevisionFixtures(editor);

    await documentFixtures.run(datasource);
    await revisionFixtures.run(datasource);

    console.log('✅ Fixtures completed!');
    process.exit(0);
  } catch (error) {
    console.error('❌ Fixtures failed:', error);
    process.exit(1);
  }
}

if (import.meta.main) {
  run().catch((error) => {
    console.error('Error during fixtures:', error);
    process.exit(1);
  });
}
