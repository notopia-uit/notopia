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

map("n", "<localleader>lrt", function()
  local clients = lsp.get_clients({ name = "tsgo" })
  for client in vim.iter(clients) do ---@cast client vim.lsp.Client
    client:stop()
  end
  lsp.start(lsp.config["gopls"])
end, { desc = "LSP | Restart TSGO", silent = true })

map("n", "<localleader>lrg", function()
  local clients = lsp.get_clients({ name = "gopls" })
  for client in vim.iter(clients) do ---@cast client vim.lsp.Client
    client:stop()
  end
  lsp.start(lsp.config["gopls"])
end, { desc = "LSP | Restart gopls", silent = true })

map("n", "<localleader>lrr", function()
  local clients = lsp.get_clients({ name = "redocly_ls" })
  for client in vim.iter(clients) do ---@cast client vim.lsp.Client
    client:stop()
  end
  lsp.start(lsp.config["redocly_ls"])
end, { desc = "LSP | Restart redocly_ls", silent = true })

vim.o.backupcopy = "yes" -- https://github.com/nrwl/nx/issues/20622
vim.opt.isfname:append("{,},@")
