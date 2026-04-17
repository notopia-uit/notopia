---
order: 1
---

# Note Database Diagram

:::info

- Only column marked with `N` mean nullable

:::

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

trashed_by: {
  shape: rectangle
  purpose
  parent
}

workspaces: {
  id: uuid {constraint: PK}
  slug: text {constraint: U}
  name: text
  created_at: timestamptz
  updated_at: timestamptz
  deleted_at: timestamptz {constraint: N}

  shape: sql_table
}

folders: {
  id: uuid {constraint: PK}
  name: text
  "icon": text {constraint: N}
  workspace_id: uuid {constraint: FK}
  parent_id: uuid {constraint: FK, N}
  created_at: timestamptz
  updated_at: timestamptz
  trashed_by: trashed_by {constraint: N}
  trashed_at: timestamptz {constraint: N}

  shape: sql_table
}

notes: {
  id: uuid {constraint: PK}
  name: text
  "icon": text {constraint: N}
  folder_id: uuid {constraint: FK}
  tags: 'text[]'
  size: integer
  created_at: timestamptz
  updated_at: timestamptz
  trashed_by: trashed_by {constraint: N}
  trashed_at: timestamptz {constraint: N}

  shape: sql_table
}

note_links: {
  source_id: uuid {constraint: PK, FK}
  target_id: uuid {constraint: PK, FK}

  shape: sql_table
}

folders.workspace_id -> workspaces.id
folders.parent_id -> folders.id
notes.folder_id -> folders.id
note_links.source_id -> notes.id
note_links.target_id -> notes.id
```

<!-- diagram id="database-diagram-note" -->
