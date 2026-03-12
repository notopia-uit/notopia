import { Block } from '@blocknote/core';
import { Injectable, NotFoundException } from '@nestjs/common';
import { randomUUID } from 'crypto';
import { Traceable } from 'nestjs-otel';

import { RevisionEntity } from './revision.entity';
import { RevisionRepository } from './revision.repository';

export interface RevisionModel {
  id: string;
  documentId: string;
  name: string | null;
  content: Block[];
  createdAt: string;
}

export interface PaginatedRevisions {
  data: RevisionModel[];
  page: number;
  limit: number;
  total: number;
}

@Injectable()
@Traceable()
export class RevisionService {
  constructor(private readonly revisionRepository: RevisionRepository) {}

  private entityToModel(entity: RevisionEntity): RevisionModel {
    return {
      id: entity.id,
      documentId: entity.documentId,
      name: entity.name,
      content: entity.content,
      createdAt: entity.createdAt.toISOString(),
    };
  }

  async getRevision(revisionId: string): Promise<RevisionModel> {
    const revision = await this.revisionRepository.getById(revisionId);
    if (!revision) {
      throw new NotFoundException(`Revision ${revisionId} not found`);
    }
    return this.entityToModel(revision);
  }

  async getRevisionsByDocumentId(
    documentId: string,
    page: number,
    limit: number
  ): Promise<PaginatedRevisions> {
    const all = await this.revisionRepository.getByDocumentId(documentId);
    const total = all.length;
    const start = (page - 1) * limit;
    const slice = all.slice(start, start + limit);
    return {
      data: slice.map((e) => this.entityToModel(e)),
      page,
      limit,
      total,
    };
  }

  async createRevision(
    documentId: string,
    content: Block[]
  ): Promise<RevisionModel> {
    const entity = new RevisionEntity();
    entity.id = randomUUID();
    entity.documentId = documentId;
    entity.name = null;
    entity.content = content;
    await this.revisionRepository.save(entity);
    return this.entityToModel(entity);
  }

  async renameRevision(revisionId: string, name: string | null): Promise<void> {
    const revision = await this.revisionRepository.getById(revisionId);
    if (!revision) {
      throw new NotFoundException(`Revision ${revisionId} not found`);
    }
    revision.name = name;
    await this.revisionRepository.save(revision);
  }

  async deleteRevision(revisionId: string): Promise<void> {
    const revision = await this.revisionRepository.getById(revisionId);
    if (!revision) {
      throw new NotFoundException(`Revision ${revisionId} not found`);
    }
    await this.revisionRepository.delete(revisionId);
  }
}
