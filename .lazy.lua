local root = vim.fn.expand("%:p:h")

---@module 'lazy'
---@type LazySpec
return {
  {
    "stevearc/conform.nvim",
    ---@module 'conform'
    ---@param opts conform.setupOpts
    ---@return conform.setupOpts
    opts = function(_, opts)
      opts.formatters_by_ft = opts.formatters_by_ft or {}
      opts.formatters_by_ft.html = nil
      opts.formatters_by_ft.css = nil
      opts.formatters_by_ft.scss = nil
      opts.formatters_by_ft.markdown = nil
      opts.formatters_by_ft.json = nil
      opts.formatters_by_ft.json5 = nil
      opts.formatters_by_ft.jsonc = nil
      opts.formatters_by_ft.toml = nil
      opts.formatters_by_ft.yaml = nil
      opts.formatters_by_ft.typescript = nil
      opts.formatters_by_ft.typescriptreact = nil
      opts.formatters_by_ft.javascriptreact = nil
      opts.formatters_by_ft.javascript = nil

      -- opts.formatters.oxlint = {
      --   command = string.format("%s/node_modules/.bin/oxlint", root),
      --   cwd = require("conform.util").root_file("oxlint.config.ts"),
      --   require_cwd = true,
      -- }
      --
      -- ---@type conform.FiletypeFormatter
      -- local formatters = {
      --   "oxlint",
      -- }
      --
      -- opts.formatters_by_ft.typescript = formatters
      -- opts.formatters_by_ft.typescriptreact = formatters
      -- opts.formatters_by_ft.javascriptreact = formatters
      -- opts.formatters_by_ft.javascript = formatters
      return opts
    end,
    optional = true,
  },

  -- {
  --   "mfussenegger/nvim-lint",
  --   opts = function(opts)
  --     local lint = require("lint")
  --
  --     lint.linters.oxlint.cmd = string.format("%s/node_modules/.bin/oxlint", root)
  --     ---@diagnostic disable-next-line: assign-type-mismatch
  --     lint.linters.oxlint.cwd = function(...)
  --       vim.notify(string.format("lint cwd args: %s", vim.inspect({ ... })), vim.log.levels.INFO)
  --       return vim.fs.root(vim.api.nvim_buf_get_name(0), "oxlint.config.ts")
  --     end
  --
  --     local js_fts = { "javascript", "javascriptreact", "typescript", "typescriptreact" }
  --
  --     for _, ft in ipairs(js_fts) do
  --       lint.linters_by_ft[ft] = {
  --         "oxlint",
  --       }
  --     end
  --     return opts
  --   end,
  --   optional = true,
  -- },

  {
    "folke/which-key.nvim",
    ---@module 'which-key'
    ---@type wk.Opts
    opts = {
      spec = {
        { "<localleader>l", group = "LSP", icon = { icon = "", color = "yellow" } },
      },
    },
    opts_extend = {
      "spec",
    },
    optional = true,
  },

  {
    "mason-org/mason-lspconfig.nvim",
    ---@module 'mason-lspconfig'
    ---@type MasonLspconfigSettings
    opts = {
      automatic_enable = false,
    },
    optional = true,
  },

  {
    "neovim/nvim-lspconfig",
  },

  {
    "windwp/nvim-ts-autotag",
    opts = {
      per_filetype = {
        typescriptreact = {
          enable_close = true,
          enable_rename = false,
          enable_close_on_slash = true,
        },
      },
    },
    optional = true,
  },
}
