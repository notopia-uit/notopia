const { NxAppWebpackPlugin } = require('@nx/webpack/app-plugin');
const { join } = require('path');

/**
 * Exporting a function defers evaluation of environment-specific values
 * (like process.env.NODE_ENV) until webpack invocation at build time,
 * rather than at NX project-graph analysis time.
 *
 * @returns {import('webpack').Configuration}
 */
module.exports = () => ({
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
    // ws optional native bindings (falls back to pure JS)
    /Can't resolve 'bufferutil'/,
    /Can't resolve 'utf-8-validate'/,
    // TypeORM optional drivers that are not installed
    /Can't resolve 'react-native-sqlite-storage'/,
    /Can't resolve '@google-cloud\/spanner'/,
    /Can't resolve 'mongodb'/,
    /Can't resolve '@sap\/hana-client(\/extension\/Stream)?'/,
    /Can't resolve 'mysql'/,
    /Can't resolve 'mysql2'/,
    /Can't resolve 'oracledb'/,
    /Can't resolve 'pg-native'/,
    /Can't resolve 'pg-query-stream'/,
    /Can't resolve 'typeorm-aurora-data-api-driver'/,
    /Can't resolve 'redis'/,
    /Can't resolve 'ioredis'/,
    /Can't resolve 'better-sqlite3'/,
    /Can't resolve 'sqlite3'/,
    /Can't resolve 'sql\.js'/,
    /Can't resolve 'mssql'/,
    // OpenTelemetry optional exporter not installed
    /Can't resolve '@opentelemetry\/exporter-jaeger'/,
    // Dynamic require in require-in-the-middle (used by OpenTelemetry)
    {
      module: /require-in-the-middle/,
      message: /Critical dependency/,
    },
    // Dynamic require in TypeORM internals (ConnectionOptionsReader, PlatformTools, etc.)
    {
      module: /typeorm/,
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
      // externalDependencies: ['pino-pretty'],
    }),
  ],
});
