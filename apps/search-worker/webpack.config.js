const { NxAppWebpackPlugin } = require('@nx/webpack/app-plugin');
const nodeExternals = require('webpack-node-externals');
const { join } = require('path');

// ESM-only packages that must be bundled by webpack (cannot be require()'d at runtime)
const ESM_PACKAGES = [/@handlewithcare\//, /@blocknote\//];

/**
 * Exporting a function defers evaluation of environment-specific values
 * (like process.env.NODE_ENV) until webpack invocation at build time,
 * rather than at NX project-graph analysis time.
 *
 * @returns {import('webpack').Configuration}
 */
module.exports = () => ({
  resolve: {
    // Enable @nx/source condition so workspace packages resolve to TypeScript source
    // files directly (via their package.json exports @nx/source field), compiled by swc.
    conditionNames: ['@nx/source', 'node', 'require', 'import', 'default'],
  },
  // Use our own nodeExternals so we can allowlist ESM-only packages for bundling.
  // externalDependencies: 'none' in NxAppWebpackPlugin disables its built-in nodeExternals.
  externals: [
    nodeExternals({
      modulesDir: join(__dirname, '../../node_modules'),
      allowlist: ESM_PACKAGES,
    }),
  ],
  output: {
    path: join(__dirname, 'dist'),
    clean: true,
    ...(process.env.NODE_ENV !== 'production' && {
      devtoolModuleFilenameTemplate: '[absolute-resource-path]',
    }),
  },
  // Suppress known optional-dependency, dynamic-require, and missing-sourcemap warnings
  ignoreWarnings: [
    // source-map-loader: missing .ts source files for packages that ship compiled output
    /Failed to parse source map/,
    // OpenTelemetry optional exporter not installed
    /Can't resolve '@opentelemetry\/exporter-jaeger'/,
    // Dynamic require in require-in-the-middle (used by OpenTelemetry)
    {
      module: /require-in-the-middle/,
      message: /Critical dependency/,
    },
    // Dynamic require in app-root-path
    {
      module: /app-root-path/,
      message: /Critical dependency/,
    },
    // Dynamic require in OpenTelemetry instrumentation
    {
      module: /@opentelemetry\/instrumentation/,
      message: /Critical dependency/,
    },
  ],
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
      // Disable the plugin's built-in nodeExternals; we provide our own above
      // with an allowlist for ESM-only packages that must be bundled.
      externalDependencies: 'none',
      mergeExternals: true,
    }),
  ],
});
