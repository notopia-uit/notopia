import { readFile, readdir } from 'fs/promises';
import * as path from 'path';
import { fileURLToPath } from 'url';

import { blocksToYDoc } from '@blocknote/core/yjs';
import { ServerBlockNoteEditor } from '@blocknote/server-util';
import { DataSource, Repository } from 'typeorm';
import { Seeder } from 'typeorm-extension';
import { encodeStateAsUpdate } from 'yjs';

import { DocumentEntity } from '#/document/document.entity';

import { parseFixtureMarkdownToBlocks } from './blocknote-fixture-transform';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

export default class DocumentFixtures implements Seeder {
  constructor(private editor: ServerBlockNoteEditor) {}

  public async run(dataSource: DataSource): Promise<void> {
    const documentRepo = dataSource.getRepository(DocumentEntity);
    const fixtureDataDir = path.join(__dirname, '../fixture-data');

    const files = (await readdir(fixtureDataDir)).filter((file) => file.endsWith('.md'));

    console.log(`Found ${files.length} markdown files to load`);

    const fixturePromises = files.map((file) =>
      this.fixtureDocument(file, fixtureDataDir, documentRepo)
    );

    await Promise.all(fixturePromises);
    console.log(`Successfully loaded ${files.length} documents`);
  }

  private async fixtureDocument(
    fileName: string,
    fixtureDataDir: string,
    documentRepo: Repository<DocumentEntity>
  ): Promise<void> {
    try {
      const filePath = path.join(fixtureDataDir, fileName);
      const content = await readFile(filePath, 'utf-8');

      const id = fileName.replace('.md', '');

      const blocks = await parseFixtureMarkdownToBlocks(this.editor, content);
      const yDoc = blocksToYDoc(this.editor.editor, blocks);
      const encodedData = encodeStateAsUpdate(yDoc);

      const document = documentRepo.create({
        id,
        data: encodedData,
        modified: false,
      });

      await documentRepo.save(document);
      console.log(`Loaded fixture document: ${id}`);
    } catch (error) {
      console.error(`Failed to load fixture document ${fileName}:`, error);
      throw error;
    }
  }
}
