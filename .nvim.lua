local lsp = vim.lsp
local map = vim.keymap.set
local root = vim.fn.getcwd()
local uri_root = vim.uri_from_fname(root)

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

lsp.config("tailwindcss", {
  filetypes = {
    "javascript",
    "javascriptreact",
    "typescript",
    "typescriptreact",
  },
  root_dir = function(bufnr, on_dir)
    local fname = vim.api.nvim_buf_get_name(bufnr)
    local allowed_paths = {
      "apps/web",
      "packages/ui",
    }
    local is_allowed = false
    for _, path in ipairs(allowed_paths) do
      if fname:match(path) then
        is_allowed = true
        break
      end
    end
    if not is_allowed then
      return
    end
    on_dir(require("lspconfig.util").root_pattern("package.json")(fname))
  end,
})

---@type lsp.vtsls
local tsgo_setting = {
  javascript = {
    format = {
      enable = false,
    },
  },
  typescript = {
    format = {
      enable = false,
    },
    preferences = {
      importModuleSpecifier = "non-relative",
    },
  },
}

lsp.config("tsgo", {
  settings = tsgo_setting,
  on_attach = function(client)
    client.server_capabilities.documentFormattingProvider = false
    client.server_capabilities.documentRangeFormattingProvider = false
  end,
})

lsp.config("oxlint", {
  cmd = function(dispatchers)
    return vim.lsp.rpc.start({ "oxlint", "--lsp" }, dispatchers)
  end,
  root_markers = { ".git" },
  init_options = {
    {
      workspaceUri = uri_root,
      options = {
        fixKind = "all",
        typeAware = true,
      },
    },
  },
  on_attach = function(client, bufnr)
    vim.api.nvim_buf_create_user_command(bufnr, "LspOxlintFixAll", function()
      client:exec_cmd({
        title = "Apply Oxlint automatic fixes",
        command = "oxc.fixAll",
        arguments = { { uri = vim.uri_from_bufnr(bufnr) } },
      })
    end, {
      desc = "Apply Oxlint automatic fixes",
    })
    vim.api.nvim_create_autocmd("BufWritePre", {
      buffer = bufnr,
      command = "LspOxlintFixAll",
    })
  end,
})

lsp.config("oxfmt", {
  cmd = function(dispatchers)
    return vim.lsp.rpc.start({ "oxfmt", "--lsp", "--config", ".oxfmtrc.jsonc" }, dispatchers)
  end,
  root_markers = { ".oxfmtrc.jsonc" },
})

lsp.enable({ "tsgo", "gopls", "oxlint", "oxfmt", "tailwindcss", "yamlls", "jsonls" })

if vim.fn.executable("harper-ls") == 1 then
  lsp.enable("harper_ls")
end

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

local function restart_lsp_client(client_name)
  local clients = lsp.get_clients({ name = client_name })
  for client in vim.iter(clients) do ---@cast client vim.lsp.Client
    client:stop()
  end
  lsp.start(lsp.config[client_name])
end

map("n", "<localleader>lrt", function()
  restart_lsp_client("tsgo")
end, { desc = "LSP | Restart TSGO", silent = true })

map("n", "<localleader>lre", function()
  restart_lsp_client("eslint")
end, { desc = "LSP | Restart eslint", silent = true })

map("n", "<localleader>lrg", function()
  restart_lsp_client("gopls")
end, { desc = "LSP | Restart gopls", silent = true })

map("n", "<localleader>lrG", function()
  restart_lsp_client("golangci_lint_ls")
end, { desc = "LSP | Restart golangci_lint_ls", silent = true })

map("n", "<localleader>lrr", function()
  restart_lsp_client("redocly_ls")
end, { desc = "LSP | Restart redocly_ls", silent = true })

vim.o.backupcopy = "yes" -- https://github.com/nrwl/nx/issues/20622
vim.opt.isfname:append("{,},@")
