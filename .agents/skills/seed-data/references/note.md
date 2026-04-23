# Note service seed

Purpose
- Seed the note service with workspace/folder/note metadata derived from the
  Obsidian vault.

Source data
- Obsidian vault: `submodule/trshpuppy-obsidian-notes/`

Workspace context
- Workspace id comes from authorization seed data:
  `internal/authorization/app/policy_test.csv`.
- The note seed uses workspace id `00000000-0000-0000-0000-000000000110` with
  slug `notopia` and name `Notopia`.

Business expectations
- A single workspace is created for the seeded vault.
- Folder hierarchy mirrors the vault directory structure.
- Root folder has an empty name and no parent.
- No icons, no trashed fields.

Notes
- Each markdown file becomes a note with:
  - name from the file basename
  - folder from the file path
  - tags parsed from Obsidian `#tag` syntax (no nested tags)
  - size derived from the JSON length of raw markdown content
- Outgoing links are created for Obsidian wiki links and markdown links that
  resolve to another vault note.

Where it lands
- SQL seed: `internal/note/seed.sql`
- Migration schema: `internal/note/infra/persistence/pgmigration/00001_init.sql`
