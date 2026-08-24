local lsp = vim.lsp
local map = vim.keymap.set
local root = vim.fn.expand("%:p:h")
local uri_root = vim.uri_from_fname(root)
local lsp_watchfile = vim.lsp._watchfiles

vim.filetype.add({
  filename = {
    ["project.json"] = "jsonc",
  },
})

-- For toggle
local enable_lsp_watchfile = false

local allowed_linked_editing_range_clients = {
  html = true,
  tsc = true,
}

lsp.config("*", {
  capabilities = {
    workspace = {
      didChangeWatchedFiles = {
        dynamicRegistration = enable_lsp_watchfile,
      },
    },
  },
} --[[@as vim.lsp.Config]])

if enable_lsp_watchfile and lsp_watchfile then
  local glob = vim.glob
  lsp_watchfile._poll_exclude_pattern = lsp_watchfile._poll_exclude_pattern
    + glob.to_lpeg(string.format("%s/submodule/**", root))
    + glob.to_lpeg(string.format("%s/deploy/**", root))
    + glob.to_lpeg(string.format("%s/http/**", root))
    + glob.to_lpeg(string.format("%s/.nx/**", root))
    + glob.to_lpeg(string.format("%s/.agent/**", root))
    + glob.to_lpeg(string.format("%s/.claude/**", root))
    + glob.to_lpeg("**/.next/**")
    + glob.to_lpeg("**/dist/**")
    + glob.to_lpeg("**/build/**")
    + glob.to_lpeg("**/__debug*")
end

vim.api.nvim_create_autocmd("LspAttach", {
  ---@param args vim.api.keyset.create_autocmd.callback_args | {data: vim.event.lspattach.data}
  callback = function(args)
    local client = lsp.get_client_by_id(args.data.client_id)
    assert(client, "LSP client not found")
    if not allowed_linked_editing_range_clients[client.name] then
      return
    end
    lsp.linked_editing_range.enable(true, {
      client_id = args.data.client_id,
    })
  end,
})

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
-- } --[[@as vim.lsp.Config]])

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
} --[[@as vim.lsp.Config]])

lsp.config("golangci_lint_ls", {
  root_dir = function(_, on_dir)
    on_dir(root)
  end,
} --[[@as vim.lsp.Config]])

lsp.config("tailwindcss", {
  cmd = { "tailwindcss-language-server", "--stdio" },
  root_dir = function(bufnr, on_dir)
    local fname = vim.api.nvim_buf_get_name(bufnr)
    local allowed_paths = {
      "apps/web",
      "apps/test-editor",
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
      files = {
        exclude = {
          "**/.git/**",
          "**/node_modules/**",
          "**/dist/**",
          "**/build/**",
        },
      },
      experimental = {
        configFile = {
          ["apps/web/app/globals.css"] = "apps/web/app/**",
          ["packages/ui/src/globals.css"] = "packages/ui/src/**",
        },
      },
    },
  },
  capabilities = {
    workspace = {
      didChangeWatchedFiles = {
        dynamicRegistration = false,
      },
    },
  },
} --[[@as vim.lsp.Config]])

lsp.config("twcssls", {
  cmd = { "mise", "exec", "npm@tailwindcss/language-server", "--", "css-language-server", "--stdio" },
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
} --[[@as vim.lsp.Config]])

lsp.config("tsc", {
  cmd = { "tsc", "--lsp", "--stdio" },
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
        importModuleSpecifier = "shortest",
        importModuleSpecifierEnding = "minimal",
      },
    },
  },
  on_attach = function(client)
    client.server_capabilities.documentFormattingProvider = false
    client.server_capabilities.documentRangeFormattingProvider = false
  end,
} --[[@as vim.lsp.Config]])

lsp.config.nestjs_doctor = {
  name = "nestjs_doctor",
  cmd = { "nestjs-doctor-lsp", "--stdio" },
  filetypes = { "typescript" },
  root_dir = function(bufnr, on_dir)
    local fname = vim.api.nvim_buf_get_name(bufnr)
    local allowed_paths = {
      "apps/document",
      "apps/search-worker",
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
  workspace_required = true,
  init_options = {
    nestjsDoctor = {
      enable = true,
      scanOnSave = true,
      scanOnOpen = true,
    },
  },
  flags = {
    debounce_text_changes = 2000,
  },
}

lsp.config("oxlint", {
  cmd = { "./node_modules/.bin/oxlint", "--lsp" },
  root_dir = function(_, on_dir)
    on_dir(root)
  end,
  settings = {
    typeAware = false,
  },
  init_options = {
    {
      workspaceUri = string.format("%s/packages/ui", uri_root),
      options = {
        configPath = string.format("%s/packages/ui/oxlint.config.ts", root),
        fixKind = "all",
        run = "onSave",
        typeAware = false,
      },
    },
    {
      workspaceUri = string.format("%s/packages/lib", uri_root),
      options = {
        configPath = string.format("%s/packages/lib/oxlint.config.ts", root),
        fixKind = "all",
        run = "onSave",
        typeAware = false,
      },
    },
    {
      workspaceUri = string.format("%s/apps/web", uri_root),
      options = {
        configPath = string.format("%s/apps/web/oxlint.config.ts", root),
        fixKind = "all",
        run = "onSave",
        typeAware = false,
      },
    },
    {
      workspaceUri = string.format("%s/apps/api-web", uri_root),
      options = {
        configPath = string.format("%s/apps/api-web/oxlint.config.ts", root),
        fixKind = "all",
        run = "onSave",
        typeAware = false,
      },
    },
    {
      workspaceUri = string.format("%s/apps/document", uri_root),
      options = {
        configPath = string.format("%s/apps/document/oxlint.config.ts", root),
        fixKind = "all",
        run = "onSave",
        typeAware = false,
      },
    },
    {
      workspaceUri = string.format("%s/apps/search-worker", uri_root),
      options = {
        configPath = string.format("%s/apps/search-worker/oxlint.config.ts", root),
        fixKind = "all",
        run = "onSave",
        typeAware = false,
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
    -- vim.api.nvim_create_autocmd("BufWritePre", {
    --   buffer = bufnr,
    --   command = "LspOxlintFixAll",
    -- })
  end,
} --[[@as vim.lsp.Config]])

lsp.config("oxfmt", {
  filetypes = {
    "css",
    "handlebars",
    "html",
    "javascript",
    "javascriptreact",
    "json",
    "json5",
    "jsonc",
    "less",
    "markdown",
    "scss",
    "toml",
    "typescript",
    "typescriptreact",
    "vue",
    "yaml",
    "yaml.github",
  },
  cmd = { "./node_modules/.bin/oxfmt", "--lsp", "--config", ".oxfmtrc.jsonc" },
  root_dir = function(_, on_dir)
    on_dir(root)
  end,
} --[[@as vim.lsp.Config]])

lsp.config("harper_ls", {
  capabilities = {
    workspace = {
      didChangeWatchedFiles = {
        dynamicRegistration = false,
        relativePatternSupport = false,
      },
    },
  },
} --[[@as vim.lsp.Config]])

lsp.enable({
  "bashls",
  "buf_ls",
  "ecfg",
  "emmet_language_server",
  "gh_actions_ls",
  "golangci_lint_ls",
  "gopls",
  "jsonls",
  "jsonls",
  "lua_ls",
  "nxls",
  "oxfmt",
  "oxlint",
  "redocly_ls",
  "tailwindcss",
  "tsc",
  "twcssls",
  "yamlls",
  "yamlls",
  -- "harper_ls",
  -- "nestjs_doctor",
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

--[[
local function track_may_progress()
  local start_date = os.time({ year = 2026, month = 5, day = 8, hour = 0 })
  local end_date = os.time({ year = 2026, month = 5, day = 31, hour = 23, min = 59 })
  local now = os.time()

  local total_duration = os.difftime(end_date, start_date)
  local seconds_in_day = 24 * 60 * 60

  local elapsed_sec = os.difftime(now, start_date)
  local remaining_sec = os.difftime(end_date, now)

  if elapsed_sec < 0 then
    elapsed_sec = 0
  end
  if remaining_sec < 0 then
    remaining_sec = 0
  end

  local days_elapsed = math.floor(elapsed_sec / seconds_in_day)
  local days_remaining = math.ceil(remaining_sec / seconds_in_day)
  local percentage = math.min(100, math.max(0, (elapsed_sec / total_duration) * 100))

  local msg =
    string.format("Passed: %d day(s)\nLeft: %d day(s)\nProgress: %.1f%%", days_elapsed, days_remaining, percentage)

  local level = vim.log.levels.INFO
  if percentage >= 99 then
    level = vim.log.levels.ERROR
    msg = msg .. "\nIt's the final countdown! We have to submit this right now"
  elseif percentage >= 98 then
    level = vim.log.levels.ERROR
    msg = msg .. "\nThe deadline is upon us! We have to submit this immediately"
  elseif percentage >= 97 then
    level = vim.log.levels.ERROR
    msg = msg .. "\nIt's the last minute! We have to submit this right now!"
  elseif percentage >= 95 then
    level = vim.log.levels.ERROR
    msg = msg .. "\nThe deadline is here! We have to submit this immediately!"
  elseif percentage >= 90 then
    level = vim.log.levels.ERROR
    msg = msg .. "\nThis is the final stretch! We have to finish this right now!"
  elseif percentage >= 85 then
    level = vim.log.levels.ERROR
    msg = msg .. "\nThe clock is ticking! We have to finish this immediately!"
  elseif percentage >= 80 then
    level = vim.log.levels.ERROR
    msg = msg .. "\nTime's running out! May is ending soon!"
  elseif percentage >= 75 then
    level = vim.log.levels.ERROR
    msg = msg .. "\nWe really need to move quickly to meet the 5 p.m. cutoff"
  elseif percentage >= 70 then
    level = vim.log.levels.ERROR
    msg = msg .. "\nWe're in the eleventh hour, so it’s crunch time to finish this"
  elseif percentage >= 60 then
    level = vim.log.levels.WARN
    msg = msg .. "\nWith the deadline looming, we have to accelerate our submission process"
  elseif percentage >= 50 then
    level = vim.log.levels.WARN
    msg = msg .. "\nWe are down to the wire on this project, so we need to wrap it up immediately"
  elseif percentage >= 40 then
    level = vim.log.levels.WARN
    msg = msg .. "\nHurry up! May is almost over!"
  end
  vim.notify(msg, level, { title = "Progress" })
end

local progress_augroup = vim.api.nvim_create_augroup("MayProgress", { clear = true })

---@type integer
local progress_ev

progress_ev = vim.api.nvim_create_autocmd("VimEnter", {
  group = progress_augroup,
  callback = function()
    vim.schedule(track_may_progress)
    vim.api.nvim_del_autocmd(progress_ev)
  end,
})
--]]
