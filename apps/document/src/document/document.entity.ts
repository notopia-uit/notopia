import { Column, Entity, OneToMany, PrimaryColumn, type Relation, UpdateDateColumn } from 'typeorm';

import { RevisionEntity } from '../revision/revision.entity';

@Entity('documents')
export class DocumentEntity {
  @PrimaryColumn('uuid', { type: 'uuid' })
  id!: string;

  @Column({ type: 'bytea' })
  data!: Buffer;

  @Column({ type: 'boolean', default: false })
  modified!: boolean;

  @Column({ type: 'timestamptz', nullable: true })
  lastCommittedAt!: Date | null;

  @UpdateDateColumn({ type: 'timestamptz' })
  updatedAt!: Date;

  @Column({ type: 'text', array: true, default: () => "'{}'" })
  pendingContributorIds!: string[];

  @OneToMany(() => RevisionEntity, (revision) => revision.document, {
    cascade: true,
  })
  revisions!: Relation<RevisionEntity[]>;
}
