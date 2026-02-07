local lsp = vim.lsp
local map = vim.keymap.set

-- lsp.config("yamlls", {
--   ---@module 'codesettings'
--   ---@type lsp.yamlls
--   settings = {
--     yaml = {
--       customTags = {
--         "!Condition sequence",
--         "!Context scalar",
--         "!Enumerate sequence",
--         "!Env scalar",
--         "!File scalar",
--         "!File sequence",
--         "!Find sequence",
--         "!Format sequence",
--         "!If sequence",
--         "!Index scalar",
--         "!KeyOf scalar",
--         "!Value scalar",
--         "!AtIndex scalar",
--       },
--     },
--   },
-- })

lsp.config("gopls", {
  settings = {
    gopls = {
      buildFlags = {
        "-tags",
        "integration",
      },
    },
  },
})

map("n", "<localleader>b", function()
  vim.ui.select({
    "none",
    "integration",
    "wireinject",
    "integration,wireinject",
  }, {
    prompt = "Select gopls build tag",
  }, function(tag)
    if not tag then
      return
    end
    local clients = lsp.get_clients({ name = "gopls" })
    for client in vim.iter(clients) do ---@cast client vim.lsp.Client
      client:stop()
    end
    lsp.config.gopls = {
      settings = {
        gopls = {
          buildFlags = tag ~= "none" and {
            "-tags",
            tag,
          } or {},
        },
      },
    }
    lsp.start(lsp.config["gopls"])
  end)
end, { desc = "LSP | Switch buildFlags", silent = true })

map("n", "<localleader>lr", function()
  local clients = lsp.get_clients({ name = "gopls" })
  for client in vim.iter(clients) do ---@cast client vim.lsp.Client
    client:stop()
  end
  lsp.start(lsp.config["gopls"])
end, { desc = "LSP | Restart gopls", silent = true })

vim.o.backupcopy = "yes" -- https://github.com/nrwl/nx/issues/20622
