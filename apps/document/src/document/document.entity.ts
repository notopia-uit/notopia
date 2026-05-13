import { Column, Entity, OneToMany, PrimaryColumn, type Relation } from 'typeorm';

import { RevisionEntity } from '#/revision/revision.entity';

@Entity('documents')
export class DocumentEntity {
  @PrimaryColumn('uuid', { type: 'uuid' })
  id!: string;

  @Column({ type: 'bytea' })
  data!: Buffer;

  @Column({ type: 'boolean', default: false })
  modified!: boolean;

  @OneToMany(() => RevisionEntity, (revision) => revision.document, {
    cascade: true,
  })
  revisions!: Relation<RevisionEntity[]>;
}
