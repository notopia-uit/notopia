import { createRequire } from 'module';
import { join } from 'path';

import { NxAppRspackPlugin } from '@nx/rspack/app-plugin.js';
import { RsdoctorRspackPlugin } from '@rsdoctor/rspack-plugin';
import type { Configuration } from '@rspack/cli';
import rspack, { type DevTool } from '@rspack/core';
// import { RunScriptWebpackPlugin } from 'run-script-webpack-plugin';

const require = createRequire(import.meta.url);
const __dirname = import.meta.dirname;

// Curated regex externals. Nx's `externalDependencies` array option only does an
// exact `ctx.request` match and hard-overwrites `config.externals`, so it can't
// externalize subpath imports. We pass `externalDependencies: 'none'` to Nx (→
// empty externals) and set a regex externals function in a trailing plugin
// (ExternalizePlugin) that runs after Nx.
//
// search-worker only talks to Kafka + Meilisearch, so the externalized set is
// small: pino + kafkajs are the OTel-instrumented packages require-in-the-middle
// must patch via a real require(). It uses no pg/grpc/@aws-sdk, so none of those
// are externalized. Everything else stays bundled — in particular ESM/interop-
// fragile packages like @blocknote/* must be compiled by rspack; externalizing
// them makes Node load their broken .cjs (e.g. "Code.extend is not a function").
//
// NOTE: keep this in sync with the dependencies/devDependencies split in
// package.json — the externalized set below must stay in `dependencies`.
const externalize: RegExp[] = [
  /^pino$/,
  /^kafkajs$/,
  /^jsdom$/, // server-util worker-file bug (#https://github.com/TypeCellOS/BlockNote/issues/1939); document app
];

class ExternalizePlugin {
  apply(compiler: rspack.Compiler) {
    // Overwrite the externals Nx set (it hard-assigns `config.externals`). Runs
    // after NxAppRspackPlugin because it sits later in the `plugins` array.
    compiler.options.externals = [
      ({ request }, callback) =>
        request && externalize.some((re) => re.test(request))
          ? callback(undefined, `commonjs ${request}`)
          : callback(),
    ];
  }
}

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
  devtool: isDev ? 'source-map' : false,
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
  plugins: [
    new NxAppRspackPlugin({
      tsConfig: tsConfigFile,
      main: 'apps/search-worker/src/main.ts',
      optimization: !isDev,
      sourceMap: devTool,
      generatePackageJson: true,
      externalDependencies: 'none',
    }),
    new ExternalizePlugin(),
    new rspack.NormalModuleReplacementPlugin(/file-type$/, require.resolve('./stub.js')),
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
  devServer: {
    allowedHosts: 'all',
  },
};

export default config;
