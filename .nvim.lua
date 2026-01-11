vim.filetype.add({
  pattern = {
    ["api/edit/.*/.*%.yaml"] = "yaml.openapi",
    ["api/note/.*/.*%.yaml"] = "yaml.openapi",
    ["api/shared/.*/.*%.yaml"] = "yaml.openapi",
  },
})
