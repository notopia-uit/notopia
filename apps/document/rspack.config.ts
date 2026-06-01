import { createRequire, isBuiltin } from 'module';
import { join } from 'path';

import { NxAppRspackPlugin } from '@nx/rspack/app-plugin.js';
import { RsdoctorRspackPlugin } from '@rsdoctor/rspack-plugin';
import type { Configuration } from '@rspack/cli';
import rspack, { type DevTool } from '@rspack/core';
// import { RunScriptWebpackPlugin } from 'run-script-webpack-plugin';

const require = createRequire(import.meta.url);
const __dirname = import.meta.dirname;

// Externals: externalize everything by default, bundle ONLY what can't be loaded
// natively. Nx's `externalDependencies` option can't express this (its array is an
// exact-match allowlist and it hard-overwrites `config.externals`), so we pass
// `externalDependencies: 'none'` to Nx (→ empty externals) and set our own
// externals function in a trailing plugin (ExternalizePlugin) that runs after Nx.
//
// Keeping deps external is what lets OpenTelemetry's require-in-the-middle patch
// them at runtime (pino → logs; pg/kafkajs/grpc/nestjs/express → spans) and keeps
// the bundle small. Blocknote's own deps (prosemirror, yjs, tiptap, the markdown
// stack, …) are externalized too — they load fine natively.
//
// What MUST be bundled is the blocknote editor cluster — these can't be loaded
// natively as external CommonJS requires, so rspack has to compile them in:
//   - @blocknote/*                 broken require()-condition build ("Code.extend
//                                  is not a function").
//   - prosemirror-* / @handlewithcare/* / @tiptap/*   ESM editor packages; some
//                                  ship no CJS "exports" main, so a native
//                                  require() throws (No "exports" main defined).
//   - yjs / y-* / lib0             CRDT identity — one shared instance.
//   - react / react-dom / scheduler / @floating-ui   blocknote's view deps.
//   - @bufbuild/protobuf           type-registry singleton (keeps inquire-shim valid).
//   - @notopia-uit/*               workspace libs — they import @blocknote, so if
//                                  left external their native require('@blocknote')
//                                  would hit the same crash.
//   - tslib / @swc/helpers         tiny transpile runtimes — always inline.
// EVERYTHING ELSE is externalized: backend infra (nestjs, rxjs, typeorm, express,
// pg, kafkajs, grpc, @aws-sdk, pino, @opentelemetry/*) AND blocknote's dual-package
// leaf deps (the unified/remark/mdast markdown stack) — they load fine natively.
// Keeping them external is what lets OpenTelemetry's require-in-the-middle patch
// pino (logs) and pg/kafkajs/grpc/nestjs/express (spans).
const keepBundled: RegExp[] = [
  /^@blocknote\//,
  /^prosemirror-/,
  /^@handlewithcare\//,
  /^@tiptap\//,
  /^yjs$/,
  /^y-protocols(\/|$)/,
  /^y-prosemirror$/,
  /^lib0(\/|$)/,
  /^react$/,
  /^react-dom(\/|$)/,
  /^scheduler(\/|$)/,
  /^@floating-ui\//,
  /^@bufbuild\/protobuf/,
  /^@hocuspocus\//,
  /^@notopia-uit\//,
  /^tslib$/,
  /^@swc\/helpers/,
];

class ExternalizePlugin {
  apply(compiler: rspack.Compiler) {
    // Overwrite the externals Nx set (it hard-assigns `config.externals`). Runs
    // after NxAppRspackPlugin because it sits later in the `plugins` array.
    compiler.options.externals = [
      ({ request }, callback) => {
        // App code, path aliases (@/, @database, #/) and node builtins → bundle.
        if (
          !request ||
          /^[./]/.test(request) ||
          request.startsWith('@/') ||
          request.startsWith('@database') ||
          request.startsWith('#/') ||
          isBuiltin(request)
        ) {
          return callback();
        }
        // @blocknote + the workspace libs that import it → bundle; rest external.
        return keepBundled.some((re) => re.test(request))
          ? callback()
          : callback(undefined, `commonjs ${request}`);
      },
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
  plugins: [
    new NxAppRspackPlugin({
      tsConfig: tsConfigFile,
      main: 'apps/document/src/main.ts',
      optimization: !isDev,
      sourceMap: devTool,
      generatePackageJson: true,
      externalDependencies: 'none',
    }),
    new ExternalizePlugin(),
    new rspack.NormalModuleReplacementPlugin(/file-type$/, require.resolve('./stub.js')),
    new rspack.NormalModuleReplacementPlugin(
      /@protobufjs\/inquire/,
      require.resolve('./inquire-shim.js')
    ),
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
    devMiddleware: {
      writeToDisk: true,
    },
  },
};

export default config;
