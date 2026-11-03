import { MigrationInterface, QueryRunner } from 'typeorm';

export class Init1793694893185 implements MigrationInterface {
  name = 'Init1793694893185';

  public async up(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(
      `CREATE TYPE "public"."revisions_trigger_enum" AS ENUM('manual', 'auto', 'sweeper')`
    );
    await queryRunner.query(
      `CREATE TABLE "revisions" ("id" uuid NOT NULL, "name" text, "content" text NOT NULL, "authorId" text, "trigger" "public"."revisions_trigger_enum" NOT NULL DEFAULT 'manual', "contributorIds" text array NOT NULL DEFAULT '{}', "contentHash" text, "createdAt" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(), "documentId" uuid, CONSTRAINT "PK_4aa9ee2c71c50508c3c501573c9" PRIMARY KEY ("id"))`
    );
    await queryRunner.query(
      `CREATE TABLE "documents" ("id" uuid NOT NULL, "data" bytea NOT NULL, "modified" boolean NOT NULL DEFAULT false, "lastCommittedAt" TIMESTAMP WITH TIME ZONE, "updatedAt" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(), "pendingContributorIds" text array NOT NULL DEFAULT '{}', CONSTRAINT "PK_ac51aa5181ee2036f5ca482857c" PRIMARY KEY ("id"))`
    );
    await queryRunner.query(
      `ALTER TABLE "revisions" ADD CONSTRAINT "FK_e5b13a890694e87fc10c2cd21cf" FOREIGN KEY ("documentId") REFERENCES "documents"("id") ON DELETE CASCADE ON UPDATE NO ACTION`
    );
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(
      `ALTER TABLE "revisions" DROP CONSTRAINT "FK_e5b13a890694e87fc10c2cd21cf"`
    );
    await queryRunner.query(`DROP TABLE "documents"`);
    await queryRunner.query(`DROP TABLE "revisions"`);
    await queryRunner.query(`DROP TYPE "public"."revisions_trigger_enum"`);
  }
}
