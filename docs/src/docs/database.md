# Database Diagram

## Note

```d2
vars: {
  d2-config: {
    layout-engine: elk
    theme-id: 3
  }
}

**.shape: sql_table
explanation.shape: rectangle

workspace: {
  id: uuid {constraint: PK}
  name: text
  root_folder_id: uuid
  created_at: timestamptz
  deleted_at: timestamptz
}

folder: {
    id: uuid {constraint: PK}
    workspace_id: uuid {constraint: FK}
    name: text
    parent_folder_id: uuid
    deleted_by_user_id: uuid
    deleted_at: timestamptz
}
```

<!-- diagram id="database-diagram-note" -->

:::info

- (FK) is `NOT NULL` by default
- Column marked with `N` mean nullable

:::

<!-- vim:set tabstop=4 shiftwidth=4: -->
