import type { Config } from 'prettier';

const config: Config = {
  singleQuote: true,
  jsxSingleQuote: false,
  trailingComma: 'es5',
  importOrderSeparation: true,
  importOrderSortSpecifiers: true,
  semi: true,
  plugins: [
    'prettier-plugin-package',
    'prettier-plugin-tailwindcss',
    '@trivago/prettier-plugin-sort-imports',
  ],
  overrides: [
    {
      files: '*.ts',
      options: {
        parser: 'typescript',
      },
    },
  ],
};

export default config;
