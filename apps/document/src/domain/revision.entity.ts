import { Column, CreateDateColumn, Entity, PrimaryColumn } from 'typeorm';

@Entity('revisions')
export class RevisionEntity {
  @PrimaryColumn('uuid')
  id!: string;

  @Column('uuid')
  documentId!: string;

  @Column({ type: 'varchar', nullable: true })
  name!: string | null;

  @Column({ type: 'bytea' })
  data!: Buffer;

  @CreateDateColumn()
  createdAt!: Date;
}
