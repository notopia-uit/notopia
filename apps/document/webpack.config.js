const { NxAppWebpackPlugin } = require('@nx/webpack/app-plugin');
const nodeExternals = require('webpack-node-externals');
const { join } = require('path');

/**
 * Exporting a function defers evaluation of environment-specific values
 * (like process.env.NODE_ENV) until webpack invocation at build time,
 * rather than at NX project-graph analysis time.
 *
 * @returns {import('webpack').Configuration}
 */
module.exports = () => ({
  resolve: {
    alias: {
      '@': join(__dirname, 'src'),
      '@database': join(__dirname, 'database'),
    },
  },
  externals: [
    nodeExternals({
      allowlist: [
        /^@notopia-uit/,
        /^@blocknote/,
        /^@handlewithcare/,
        /^prosemirror/,
      ],
    }),
  ],
  output: {
    path: join(__dirname, 'dist'),
  },
  plugins: [
    new NxAppWebpackPlugin({
      target: 'node',
      compiler: 'swc',
      main: './src/main.ts',
      tsConfig: './tsconfig.app.json',
      assets: ['./src/assets'],
      outputHashing: process.env['NODE_ENV'] === 'production' ? 'all' : 'none',
      optimization: process.env['NODE_ENV'] === 'production',
      generatePackageJson: false,
      sourceMap: true,
      externalDependencies: 'none',
      mergeExternals: true,
    }),
  ],
});
