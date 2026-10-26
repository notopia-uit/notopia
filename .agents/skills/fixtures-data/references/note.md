# Note service fixtures

Purpose
- Load fixtures into the note service with workspace/folder/note metadata derived from the
  Obsidian vault.

Source data
- Obsidian vault: `submodule/trshpuppy-obsidian-notes/`

Workspace context
- Workspace id comes from authorization fixture data:
  `internal/authorization/app/policy_fixtures.csv`.
- The note fixtures use workspace id `00000000-0000-0000-0000-000000000110` with
  slug `notopia` and name `Notopia`.

Business expectations
- A single workspace is created for the vault fixtures.
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
- SQL fixtures: `internal/notefixturegen/fixtures.sql` and generated `fixtures.gen.sql`
- Migration schema: `internal/note/infra/persistence/pgmigration/00001_init.sql`
