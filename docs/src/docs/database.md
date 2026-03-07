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
    theme-id: 3
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
  created_at: timestamptz
  updated_at: timestamptz
  deleted_by: deleted_by
  deleted_at: timestamptz {constraint: N}

  shape: sql_table
}

tags: {
  id: uuid {constraint: PK}
  name: text
  workspace_id: uuid {constraint: FK}
  created_at: timestamptz

  shape: sql_table
}

note_tags: {
  note_id: uuid {constraint: PK, FK}
  tag_id: uuid {constraint: PK, FK}

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
notes.current_revision_id -> revisions.id
tags.workspace_id -> workspaces.id
note_tags.note_id -> notes.id
note_tags.tag_id -> tags.id
revisions.note_id -> notes.id
```

<!-- diagram id="database-diagram-note" -->

## Document

```d2
vars: {
  d2-config: {
    layout-engine: elk
    theme-id: 3
  }
}

documents: {
  id: uuid {constraint: PK}
  data: bytea

  shape: sql_table
}
```
