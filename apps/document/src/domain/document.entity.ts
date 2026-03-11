import { Column, Entity, PrimaryColumn } from 'typeorm';

@Entity('documents')
export class DocumentEntity {
  @PrimaryColumn('uuid')
  id!: string;

  @Column({ type: 'bytea' })
  data!: Buffer;
}
