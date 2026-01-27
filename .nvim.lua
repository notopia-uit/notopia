vim.lsp.config("yamlls", {
  ---@module 'codesettings'
  ---@type lsp.yamlls
  settings = {
    yaml = {
      customTags = {
        "!Condition sequence",
        "!Context scalar",
        "!Enumerate sequence",
        "!Env scalar",
        "!File scalar",
        "!File sequence",
        "!Find sequence",
        "!Format sequence",
        "!If sequence",
        "!Index scalar",
        "!KeyOf scalar",
        "!Value scalar",
        "!AtIndex scalar",
      },
    },
  },
})

vim.o.backupcopy = "yes" -- https://github.com/nrwl/nx/issues/20622
