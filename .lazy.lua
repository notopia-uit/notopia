---@module 'lazy'
---@type LazySpec
return {
  "stevearc/conform.nvim",
  ---@module 'conform'
  ---@param opts conform.setupOpts
  ---@return conform.setupOpts
  opts = function(_, opts)
    opts.formatters.prettier = {
      command = function(_, ctx)
        local root = vim.fs.root(ctx.dirname, { ".git", "pnpm-workspace.yaml" })
        if root then
          local root_prettier = root .. "/node_modules/.bin/prettier"
          if vim.fn.executable(root_prettier) == 1 then
            return root_prettier
          end
        end
        return "prettier"
      end,
      cwd = require("conform.util").root_file({ "pnpm-workspace.yaml" }),
    }
    return opts
  end,
  optional = true,
}
