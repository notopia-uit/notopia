# Database Diagram

:::info

- (FK) is `NOT NULL` by default
- Column marked with `N` mean nullable

:::

## Note

```d2
vars: {
  d2-config: {
    layout-engine: elk
    theme-overrides: {
      N1: "#4c4f69"
      N2: "#5c5f77"
      N4: "#acb0be"
      N5: "#ccd0da"
      N7: "#eff1f5"
      B1: "#4c4f69"
      B2: "#6c6f85"
      B3: "#bcc0cc"
      B4: "#ccd0da"
      B5: "#dce0e8"
      B6: "#eff1f5"
      AA4: "#1e66f5"
      AA5: "#7287fd"
      AB4: "#8839ef"
      AB5: "#dc8a78"
    }
  }
}

deleted_by: {
  shape: rectangle
  purpose
  parent
}

workspaces: {
  id: uuid {constraint: PK}
  name: text
  root_folder_id: uuid
  created_at: timestamptz
  updated_at: timestamptz
  deleted_at: timestamptz {constraint: N}

  shape: sql_table
}

folders: {
  id: uuid {constraint: PK}
  name: text
  workspace_id: uuid {constraint: FK}
  parent_id: uuid
  created_at: timestamptz
  updated_at: timestamptz
  deleted_by: deleted_by
  deleted_at: timestamptz {constraint: N}

  shape: sql_table
}

notes: {
  id: uuid {constraint: PK}
  name: text
  folder_id: uuid {constraint: FK}
  current_revision_id: uuid {constraint: FK}
  tags: 'text[]'
  created_at: timestamptz
  updated_at: timestamptz
  deleted_by: deleted_by
  deleted_at: timestamptz {constraint: N}

  shape: sql_table
}

note_links: {
  source_id: uuid {constraint: PK, FK}
  target_id: uuid {constraint: PK, FK}

  shape: sql_table
}

revisions: {
  id: uuid {constraint: PK}
  note_id: uuid {constraint: FK}
  name: text {constraint: N}
  block_note_content: text
  created_at: timestamptz
  updated_at: timestamptz
  deleted_at: timestamptz {constraint: N}

  shape: sql_table
}

folders.workspace_id -> workspaces.id
folders.parent_id -> folders.id
notes.folder_id -> folders.id
note_links.source_id -> notes.id
note_links.target_id -> notes.id
notes.current_revision_id -> revisions.id
revisions.note_id -> notes.id
```

<!-- diagram id="database-diagram-note" -->

## Document

```d2
vars: {
  d2-config: {
    layout-engine: elk
    theme-overrides: {
      N1: "#4c4f69"
      N2: "#5c5f77"
      N4: "#acb0be"
      N5: "#ccd0da"
      N7: "#eff1f5"
      B1: "#4c4f69"
      B2: "#6c6f85"
      B3: "#bcc0cc"
      B4: "#ccd0da"
      B5: "#dce0e8"
      B6: "#eff1f5"
      AA4: "#1e66f5"
      AA5: "#7287fd"
      AB4: "#8839ef"
      AB5: "#dc8a78"
    }
  }
}

documents: {
  id: uuid {constraint: PK}
  data: bytea

  shape: sql_table
}
```
