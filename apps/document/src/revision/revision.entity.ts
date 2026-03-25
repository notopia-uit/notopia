import { DocumentEntity } from '../document/document.entity';
import { Block } from '@blocknote/core';
import {
  Column,
  CreateDateColumn,
  DeleteDateColumn,
  Entity,
  ManyToOne,
  PrimaryColumn,
} from 'typeorm';

@Entity('revisions')
export class RevisionEntity {
  @PrimaryColumn('uuid')
  id!: string;

  @ManyToOne(() => DocumentEntity, (document) => document.revisions, {
    onDelete: 'CASCADE',
  })
  document!: DocumentEntity;

  @Column({ type: 'text', nullable: true })
  name!: string | null;

  @Column({ type: 'simple-json' })
  content!: Block[];

  @CreateDateColumn()
  createdAt!: Date;

  @DeleteDateColumn()
  deletedAt!: Date | null;
}
