import type { Config } from 'prettier';

const config: Config = {
  singleQuote: true,
  jsxSingleQuote: false,
  trailingComma: 'es5',
  semi: true,
  plugins: ['prettier-plugin-package', 'prettier-plugin-tailwindcss'],
};

export default config;
