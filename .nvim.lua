local lsp = vim.lsp
local map = vim.keymap.set
local root = vim.fn.getcwd()
local uri_root = vim.uri_from_fname(root)

-- lsp.config("yamlls", {
--   ---@module 'lspconfig'
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
  ---@module 'lspconfig'
  ---@type lspconfig.settings.gopls
  settings = {
    gopls = {
      buildFlags = {
        "-tags",
        "integration",
      },
    },
  },
  root_dir = function(_, on_dir)
    on_dir(root)
  end,
})

lsp.config("tailwindcss", {
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
    on_dir(root)
  end,
  ---@module 'lspconfig'
  ---@type lspconfig.settings.tailwindcss
  settings = {
    tailwindCSS = {
      experimental = {
        configFile = {
          ["apps/web/app/globals.css"] = "apps/web/app/**",
          ["packages/ui/src/globals.css"] = "packages/ui/src/**",
        },
      },
    },
  },
})

lsp.config("twcssls", {
  cmd = { "css-language-server", "--stdio" },
  root_dir = function(_, on_dir)
    on_dir(root)
  end,
  workspace_folders = {
    {
      name = "web",
      uri = string.format("%s/apps/web", uri_root),
    },
    {
      name = "ui",
      uri = string.format("%s/packages/ui", uri_root),
    },
  },
})

lsp.config("tsgo", {
  cmd = { "tsgo", "--lsp", "--stdio" },
  root_dir = function(_, on_dir)
    on_dir(root)
  end,
  ---@type lspconfig.settings.vtsls
  settings = {
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
  },
  on_attach = function(client)
    client.server_capabilities.documentFormattingProvider = false
    client.server_capabilities.documentRangeFormattingProvider = false
  end,
})

lsp.config.nestjs_doctor = {
  name = "nestjs_doctor",
  cmd = { "nestjs-doctor-lsp" },
  root_dir = function(_, on_dir)
    on_dir(root)
  end,
  workspace_required = true,
  init_options = {
    nestjsDoctor = {
      enable = true,
      scanOnSave = true,
      scanOnOpen = true,
      debounceMs = 200,
    },
  },
}

lsp.config("oxlint", {
  cmd = { "oxlint", "--lsp" },
  root_dir = function(_, on_dir)
    on_dir(root)
  end,
  init_options = {
    {
      workspaceUri = string.format("%s/packages/ui", uri_root),
      options = {
        fixKind = "all",
      },
    },
    {
      workspaceUri = string.format("%s/packages/lib", uri_root),
      options = {
        fixKind = "all",
      },
    },
    {
      workspaceUri = string.format("%s/apps/web", uri_root),
      options = {
        fixKind = "all",
      },
    },
    {
      workspaceUri = string.format("%s/apps/api-web", uri_root),
      options = {
        fixKind = "all",
      },
    },
    {
      workspaceUri = string.format("%s/apps/document", uri_root),
      options = {
        fixKind = "all",
      },
    },
    {
      workspaceUri = string.format("%s/apps/search-worker", uri_root),
      options = {
        fixKind = "all",
      },
    },
  },
  workspace_folders = {
    {
      name = "ui",
      uri = string.format("%s/packages/ui", uri_root),
    },
    {
      name = "lib",
      uri = string.format("%s/packages/lib", uri_root),
    },
    {
      name = "web",
      uri = string.format("%s/apps/web", uri_root),
    },
    {
      name = "api-web",
      uri = string.format("%s/apps/api-web", uri_root),
    },
    {
      name = "document",
      uri = string.format("%s/apps/document", uri_root),
    },
    {
      name = "search-worker",
      uri = string.format("%s/apps/search-worker", uri_root),
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
  cmd = { "oxfmt", "--lsp", "--config", ".oxfmtrc.jsonc" },
  root_dir = function(_, on_dir)
    on_dir(root)
  end,
})

lsp.enable({
  "ecfg",
  "gh_actions_ls",
  "golangci_lint_ls",
  "gopls",
  "harper_ls",
  "jsonls",
  "jsonls",
  "lua_ls",
  "nestjs_doctor",
  "nxls",
  "oxfmt",
  "oxlint",
  "redocly_ls",
  "tailwindcss",
  "tsgo",
  "twcssls",
  "yamlls",
  "yamlls",
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

local function restart_lsp_client(client_name)
  local clients = lsp.get_clients({ name = client_name, bufnr = 0 })
  for client in vim.iter(clients) do ---@cast client vim.lsp.Client
    client:stop()
  end
  lsp.start(lsp.config[client_name])
end

map("n", "<localleader>lrt", function()
  restart_lsp_client("tsgo")
end, { desc = "LSP | Restart tsgo", silent = true })

map("n", "<localleader>lrc", function()
  restart_lsp_client("twcssls")
end, { desc = "LSP | Restart twcssls", silent = true })

map("n", "<localleader>lro", function()
  restart_lsp_client("oxlint")
end, { desc = "LSP | Restart oxlint", silent = true })

map("n", "<localleader>lrO", function()
  restart_lsp_client("oxfmt")
end, { desc = "LSP | Restart oxfmt", silent = true })

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
