# Document service fixtures

Purpose
- Load fixtures into the document service with parsed BlockNote content and revisions derived
  from the Obsidian vault.

Source data
- Obsidian vault: `submodule/trshpuppy-obsidian-notes/`
- Intermediate markdown: `apps/document/database/fixture-data/`

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
- Transform tool and fixture seeders live under `apps/document/database/`.
- Fixtures use BlockNote server utilities and Yjs.
