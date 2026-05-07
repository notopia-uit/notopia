import { readFile, readdir } from 'fs/promises';
import * as path from 'path';
import { fileURLToPath } from 'url';

import { blocksToYDoc } from '@blocknote/core/yjs';
import { ServerBlockNoteEditor } from '@blocknote/server-util';
import { DataSource, Repository } from 'typeorm';
import { Seeder } from 'typeorm-extension';
import { encodeStateAsUpdateV2 } from 'yjs';

import { DocumentEntity } from '#/document/document.entity';

import { parseSeedMarkdownToBlocks } from './blocknote-seed-transform';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

export default class DocumentSeeder implements Seeder {
  constructor(private editor: ServerBlockNoteEditor) {}

  public async run(dataSource: DataSource): Promise<void> {
    const documentRepo = dataSource.getRepository(DocumentEntity);
    const seedDataDir = path.join(__dirname, '../seed-data');

    const files = (await readdir(seedDataDir)).filter((file) => file.endsWith('.md'));

    console.log(`Found ${files.length} markdown files to seed`);

    const seedPromises = files.map((file) => this.seedDocument(file, seedDataDir, documentRepo));

    await Promise.all(seedPromises);
    console.log(`Successfully seeded ${files.length} documents`);
  }

  private async seedDocument(
    fileName: string,
    seedDataDir: string,
    documentRepo: Repository<DocumentEntity>
  ): Promise<void> {
    try {
      const filePath = path.join(seedDataDir, fileName);
      const content = await readFile(filePath, 'utf-8');

      const id = fileName.replace('.md', '');

      const blocks = await parseSeedMarkdownToBlocks(this.editor, content);
      const yDoc = blocksToYDoc(this.editor.editor, blocks);
      const encodedData = encodeStateAsUpdateV2(yDoc);

      const document = documentRepo.create({
        id,
        data: encodedData,
        modified: false,
      });

      await documentRepo.save(document);
      console.log(`Seeded document: ${id}`);
    } catch (error) {
      console.error(`Failed to seed document ${fileName}:`, error);
      throw error;
    }
  }
}
