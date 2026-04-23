import { ServerBlockNoteEditor } from '@blocknote/server-util';
import * as fs from 'fs';
import * as path from 'path';
import { DataSource, Repository } from 'typeorm';
import { Seeder } from 'typeorm-extension';
import { fileURLToPath } from 'url';

import { DocumentEntity } from '#/document/document.entity';
import { RevisionEntity } from '#/revision/revision.entity';

import { parseSeedMarkdownToBlocks } from './blocknote-seed-transform';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

export default class RevisionSeeder implements Seeder {
  constructor(private editor: ServerBlockNoteEditor) {}

  public async run(dataSource: DataSource): Promise<void> {
    const documentRepo = dataSource.getRepository(DocumentEntity);
    const revisionRepo = dataSource.getRepository(RevisionEntity);
    const seedDataDir = path.join(__dirname, '../seed-data');

    const files = fs
      .readdirSync(seedDataDir)
      .filter((file) => file.endsWith('.md'));

    console.log(`Found ${files.length} markdown files to create revisions for`);

    const revisionPromises = files.map((file) =>
      this.createRevision(file, seedDataDir, documentRepo, revisionRepo)
    );

    await Promise.all(revisionPromises);
    console.log(`Successfully created ${files.length} revisions`);
  }

  private async createRevision(
    fileName: string,
    seedDataDir: string,
    documentRepo: Repository<DocumentEntity>,
    revisionRepo: Repository<RevisionEntity>
  ): Promise<void> {
    try {
      const filePath = path.join(seedDataDir, fileName);
      const content = fs.readFileSync(filePath, 'utf-8');

      const documentId = fileName.replace('.md', '');

      const document = await documentRepo.findOneBy({ id: documentId });
      if (!document) {
        console.warn(
          `Document ${documentId} not found, skipping revision creation`
        );
        return;
      }

      const blocks = await parseSeedMarkdownToBlocks(this.editor, content);

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
