import { MyBlock } from '@blocknote/core';
import {
  Column,
  CreateDateColumn,
  DeleteDateColumn,
  Entity,
  ManyToOne,
  PrimaryColumn,
  type Relation,
} from 'typeorm';

import { DocumentEntity } from '#/document/document.entity';

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

  @CreateDateColumn({ type: 'timestamptz' })
  createdAt!: Date;

  @DeleteDateColumn({ type: 'timestamptz' })
  deletedAt!: Date | null;
}
