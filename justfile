all parallel="4" exclude="tag:scope:docs":
    pnpm exec nx run-many -t lint build gen bundle --parallel={{ parallel }} --exclude={{ exclude }}
