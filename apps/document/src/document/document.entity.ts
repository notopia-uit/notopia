import { RevisionEntity } from '../revision/revision.entity';
import { Column, Entity, OneToMany, PrimaryColumn } from 'typeorm';

@Entity('documents')
export class DocumentEntity {
  @PrimaryColumn('uuid')
  id!: string;

  @Column({ type: 'bytea' })
  data!: Buffer;

  @Column({ type: 'boolean', default: false })
  modified!: boolean;

  @OneToMany(() => RevisionEntity, (revision) => revision.document, {
    cascade: true,
  })
  revisions!: RevisionEntity[];
}
