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
      opts.formatters_by_ft.javascript = nil
      opts.formatters_by_ft.json = nil
      opts.formatters_by_ft.json5 = nil
      opts.formatters_by_ft.jsonc = nil
      opts.formatters_by_ft.typescript = nil
      opts.formatters_by_ft.typescriptreact = nil
      opts.formatters_by_ft.javascriptreact = nil
      opts.formatters_by_ft.toml = nil
      opts.formatters_by_ft.yaml = nil
      return opts
    end,
    optional = true,
  },
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
      automatic_enable = {
        exclude = { "eslint" },
      },
    },
    optional = true,
  },
}
