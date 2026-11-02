---
order: 3
---

# Authorization Database Diagram

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

casbin_rules: {
  id: uuid {constraint: PK}
  ptype: text
  v0: text
  v1: text
  v2: text
  v3: text {constraint: N}
  v4: text {constraint: N}
  v5: text {constraint: N}

  shape: sql_table
}
```

<!-- diagram id="database-diagram-authorization" -->
