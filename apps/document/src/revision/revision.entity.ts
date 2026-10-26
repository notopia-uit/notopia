import { MyBlock } from '@blocknote/core';
import { Column, CreateDateColumn, Entity, ManyToOne, PrimaryColumn, type Relation } from 'typeorm';

import { REVISION_TRIGGERS, type RevisionTrigger } from '../document/commit-policy';
import { DocumentEntity } from '../document/document.entity';

@Entity('revisions')
export class RevisionEntity {
  @PrimaryColumn('uuid', { type: 'uuid' })
  id!: string;

  @ManyToOne(() => DocumentEntity, (document) => document.revisions, {
    onDelete: 'CASCADE',
  })
  document!: Relation<DocumentEntity>;

  @Column({ type: 'text', nullable: true })
  name!: string | null;

  @Column({ type: 'simple-json' })
  content!: MyBlock[];

  @Column({ type: 'text', nullable: true })
  authorId!: string | null;

  @Column({ type: 'enum', enum: REVISION_TRIGGERS, default: 'manual' })
  trigger!: RevisionTrigger;

  @Column({ type: 'text', array: true, default: () => "'{}'" })
  contributorIds!: string[];

  @Column({ type: 'text', nullable: true })
  contentHash!: string | null;

  @CreateDateColumn({ type: 'timestamptz' })
  createdAt!: Date;
}
