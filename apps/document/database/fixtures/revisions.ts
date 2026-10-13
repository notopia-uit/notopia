import { readFile, readdir } from 'fs/promises';
import * as path from 'path';
import { fileURLToPath } from 'url';

import { ServerBlockNoteEditor } from '@blocknote/server-util';
import { DataSource, Repository } from 'typeorm';
import { Seeder } from 'typeorm-extension';

import { DocumentEntity } from '#/document/document.entity';
import { RevisionEntity } from '#/revision/revision.entity';

import { parseFixtureMarkdownToBlocks } from './blocknote-fixture-transform';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

export default class RevisionFixtures implements Seeder {
  constructor(private editor: ServerBlockNoteEditor) {}

  public async run(dataSource: DataSource): Promise<void> {
    const documentRepo = dataSource.getRepository(DocumentEntity);
    const revisionRepo = dataSource.getRepository(RevisionEntity);
    const fixtureDataDir = path.join(__dirname, '../fixture-data');

    const files = (await readdir(fixtureDataDir)).filter((file) => file.endsWith('.md'));

    console.log(`Found ${files.length} markdown files to create revisions for`);

    const revisionPromises = files.map((file) =>
      this.createRevision(file, fixtureDataDir, documentRepo, revisionRepo)
    );

    await Promise.all(revisionPromises);
    console.log(`Successfully created ${files.length} revisions`);
  }

  private async createRevision(
    fileName: string,
    fixtureDataDir: string,
    documentRepo: Repository<DocumentEntity>,
    revisionRepo: Repository<RevisionEntity>
  ): Promise<void> {
    try {
      const filePath = path.join(fixtureDataDir, fileName);
      const content = await readFile(filePath, 'utf-8');

      const documentId = fileName.replace('.md', '');

      const document = await documentRepo.findOneBy({ id: documentId });
      if (!document) {
        console.warn(`Document ${documentId} not found, skipping revision creation`);
        return;
      }

      const blocks = await parseFixtureMarkdownToBlocks(this.editor, content);

      const revision = revisionRepo.create({
        id: documentId,
        document,
        content: blocks,
        name: null,
      });

      await revisionRepo.save(revision);
      console.log(`Created revision for document: ${documentId}`);
    } catch (error) {
      console.error(`Failed to create revision for ${fileName}:`, error);
      throw error;
    }
  }
}
