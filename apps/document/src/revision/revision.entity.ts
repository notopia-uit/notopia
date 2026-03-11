import { Block } from '@blocknote/core';
import {
  Column,
  CreateDateColumn,
  DeleteDateColumn,
  Entity,
  PrimaryColumn,
} from 'typeorm';

@Entity('revisions')
export class RevisionEntity {
  @PrimaryColumn('uuid')
  id!: string;

  @Column('uuid')
  documentId!: string;

  @Column({ type: 'text', nullable: true })
  name!: string | null;

  @Column({ type: 'simple-json' })
  content!: Block[];

  @CreateDateColumn()
  createdAt!: Date;

  @DeleteDateColumn()
  deletedAt!: Date | null;
}
