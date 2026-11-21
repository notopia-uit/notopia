---
order: 3
---

# Document Database Diagram

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

documents: {
  id: uuid {constraint: PK}
  data: bytea
  modified: boolean
  lastCommittedAt: timestamptz {constraint: N}
  updatedAt: timestamptz
  pendingContributorIds: text[]

  shape: sql_table
}

revisions: {
  id: uuid {constraint: PK}
  name: text {constraint: N}
  content: text
  documentId: uuid {constraint: FK, N}
  authorId: text {constraint: N}
  trigger: enum
  contributorIds: text[]
  contentHash: text {constraint: N}
  createdAt: timestamptz

  shape: sql_table
}

revisions.documentId -> documents.id
```

<!-- diagram id="database-diagram-document" -->

Revision content is serialized JSON stored in a text column. Revisions are hard-deleted, and
deleting a document cascades to its revisions. See [flows](./flows) for commit and deletion behavior.
