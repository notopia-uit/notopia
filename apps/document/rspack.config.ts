import { builtinModules, createRequire } from 'module';
import { join } from 'path';

import { NxAppRspackPlugin } from '@nx/rspack/app-plugin.js';
import { RsdoctorRspackPlugin } from '@rsdoctor/rspack-plugin';
import type { Configuration } from '@rspack/cli';
import rspack, { type DevTool } from '@rspack/core';
// import { RunScriptWebpackPlugin } from 'run-script-webpack-plugin';
import nodeExternals from 'webpack-node-externals';

const require = createRequire(import.meta.url);
const __dirname = import.meta.dirname;

const isEsm = false;
const tsConfigFile = join(__dirname, 'tsconfig.app.json');

const NODE_ENV = process.env['NODE_ENV'] || 'development';
const isDev = NODE_ENV === 'development';
const useHmr = false; // because nx handle that. But it isn't nicely I guess, it stop the whole process

const devTool: DevTool = isDev ? 'source-map' : false;
const lazyImports = new Set([
  '@aws-sdk/credential-providers',
  '@google-cloud/spanner',
  '@mongodb-js/zstd',
  '@nestjs/core',
  '@nestjs/microservices',
  '@nestjs/microservices/microservices-module',
  '@nestjs/platform-express',
  '@nestjs/websockets',
  '@nestjs/websockets/socket-module',
  '@opentelemetry/winston-transport',
  '@sap/hana-client',
  '@sap/hana-client/extension/Stream',
  'amqp-connection-manager',
  'amqplib',
  'aws4',
  'better-sqlite3',
  'bson-ext',
  'bufferutil',
  'cache-manager',
  'canvas',
  'class-transformer',
  'class-validator',
  'file-type',
  'gcp-metadata',
  'hiredis',
  'ioredis',
  'kafkajs',
  'kerberos',
  'long',
  'mongodb',
  'mongodb-client-encryption',
  'mqtt',
  'mssql',
  'mysql',
  'mysql2',
  'nats',
  'oracledb',
  'pg-native',
  'pg-query-stream',
  'react-native-sqlite-storage',
  'redis',
  'snappy',
  'snappy/package.json',
  'socket.io-adapter',
  'spanner',
  'sql.js',
  'sqlite3',
  'typeorm-aurora-data-api-driver',
  'utf-8-validate',
]);

const config: Configuration = {
  target: 'node',
  devtool: devTool,
  entry: useHmr ? ['@rspack/core/hot/poll?1000', './src/main.ts'] : ['./src/main.ts'],
  output: {
    module: isEsm,
    path: join(__dirname, './dist'),
    chunkFormat: isEsm ? 'module' : 'array-push',
    filename: `[name].${isEsm ? 'm' : 'c'}js`,
    chunkFilename: '[name].[contenthash].js',
  },
  experiments: {
    outputModule: isEsm,
  },
  externalsPresets: { node: true },
  optimization: {
    minimizer: [
      new rspack.SwcJsMinimizerRspackPlugin({
        minimizerOptions: {
          compress: {
            keep_classnames: true,
            keep_fnames: true,
          },
          mangle: {
            keep_classnames: true,
            keep_fnames: true,
          },
        },
      }),
    ],
  },
  module: {
    rules: [
      {
        test: /\.tsx?$/,
        use: [
          {
            loader: 'builtin:swc-loader',
            options: {
              detectSyntax: 'auto',
              jsc: {
                parser: {
                  syntax: 'typescript',
                  decorators: true,
                },
                transform: {
                  legacyDecorator: true,
                  decoratorMetadata: true,
                },
                target: 'esnext',
              },
            },
          },
        ],
        exclude: /node_modules/,
        type: isEsm ? ('javascript/esm' as const) : ('javascript/auto' as const),
      },
      {
        test: /\.node$/,
        use: [
          {
            loader: 'node-loader',
            options: {
              name: '[path][name].[ext]',
            },
          },
        ],
      },
    ],
  },
  resolve: {
    extensions: ['.tsx', '.ts', '.js'],
    extensionAlias: isEsm
      ? {
          '.js': ['.ts', '.js'],
          '.mjs': ['.mts', '.mjs'],
        }
      : undefined,
    tsConfig: tsConfigFile,
    alias: {
      '@': join(__dirname, 'src'),
      '@database': join(__dirname, 'database'),
    },
  },
  externals: [
    // nodeExternals returns an untyped function that webpack expects
    nodeExternals({
      importType: isEsm ? 'module' : 'commonjs',
      allowlist: [/@rspack\/core\/hot\/poll/],
    }) as unknown as Configuration['externals'],
    ...(isEsm
      ? [
          ((data: any, callback: any) => {
            const request: string | undefined = data.request;
            if (!request) return callback(null);
            const bare = request.startsWith('node:') ? request.slice(5) : request;
            if (builtinModules.includes(bare)) {
              return callback(null, `node:${bare}`);
            }
            callback(null);
          }) as unknown as Configuration['externals'],
        ]
      : []),
  ] as Configuration['externals'],
  plugins: [
    new NxAppRspackPlugin({
      tsConfig: tsConfigFile,
      main: 'apps/document/src/main.ts',
      optimization: !isDev,
      sourceMap: devTool,
      generatePackageJson: true,
    }),
    new rspack.CopyRspackPlugin({
      patterns: [{ from: join(__dirname, '../../proto'), to: 'proto' }],
    }),
    new rspack.NormalModuleReplacementPlugin(/file-type$/, require.resolve('./stub.js')),
    new rspack.NormalModuleReplacementPlugin(
      /@protobufjs\/inquire/,
      require.resolve('./inquire-shim.js')
    ),
    new rspack.SourceMapDevToolPlugin({}),
    new rspack.IgnorePlugin({
      checkResource(resource) {
        if (!lazyImports.has(resource)) {
          return false;
        }
        try {
          require.resolve(resource, { paths: [process.cwd()] });
          return false;
        } catch {
          return true;
        }
      },
    }),
    useHmr && new rspack.HotModuleReplacementPlugin(),
    // isDev && new RunScriptWebpackPlugin({ name: 'main.cjs', autoRestart: false }),
    process.env.RSDOCTOR === 'true' ? new RsdoctorRspackPlugin() : undefined,
  ],
  ignoreWarnings: [
    // Dynamic require(expression) in third-party packages — structural, cannot be fixed externally
    { module: /typeorm/, message: /Critical dependency/ },
    { module: /nestjs/, message: /Critical dependency/ },
    { module: /express/, message: /Critical dependency/ },
    { module: /app-root-path/, message: /Critical dependency/ },
    { module: /load-esm/, message: /Critical dependency/ },
  ],
  devServer: {
    allowedHosts: 'all',
  },
};

export default config;
