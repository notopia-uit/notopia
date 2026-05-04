# Document service seed

Purpose
- Seed the document service with parsed BlockNote content and revisions derived
  from the Obsidian vault.

Source data
- Obsidian vault: `submodule/trshpuppy-obsidian-notes/`
- Intermediate markdown: `apps/document/database/seed-data/`

UUID mapping
- Use deterministic UUIDs derived from vault-relative paths without `.md`.
- This mapping must match the note service UUIDs so documents and notes align.

Business expectations
- Each vault note becomes:
  - One document row (binary Yjs state)
  - One revision row (parsed BlockNote blocks)
- Revisions attach to documents using the same UUID.

Where it lands
- Document entity: `apps/document/src/document/document.entity.ts`
- Revision entity: `apps/document/src/revision/revision.entity.ts`

Transform behavior
- Strip frontmatter before parsing.
- Replace Obsidian links/tags with custom inline anchors so BlockNote parsing
  preserves references and tags.
- Parse markdown to HTML, then HTML to BlockNote blocks to keep custom inline
  nodes intact.

Operational references
- Transform tool and seeders live under `apps/document/database/`.
- Seeding uses BlockNote server utilities and Yjs.
