import { builtinModules } from 'module';
import { join } from 'path';

import { NxAppRspackPlugin } from '@nx/rspack/app-plugin';
import type { Configuration } from '@rspack/cli';
import rspack from '@rspack/core';
import { RunScriptWebpackPlugin } from 'run-script-webpack-plugin';
import nodeExternals from 'webpack-node-externals';

const isEsm = false;
const tsConfigFile = join(__dirname, 'tsconfig.app.json');

const NODE_ENV = process.env['NODE_ENV'] || 'development';
const isDev = NODE_ENV === 'development';

const config: Configuration = {
  target: 'node',
  devtool: NODE_ENV === 'production' ? false : 'source-map',
  entry: isDev ? ['@rspack/core/hot/poll?1000', './src/main.ts'] : ['./src/main.ts'],
  output: {
    module: isEsm,
    path: join(__dirname, './dist'),
    chunkFormat: isEsm ? 'module' : 'array-push',
    filename: `[name].${isEsm ? 'm' : 'c'}js`,
    chunkFilename: '[name].[contenthash].js',
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
      sourceMap: NODE_ENV !== 'production',
      optimization: NODE_ENV === 'production',
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
        const lazyImports = [
          '@aws-sdk/credential-providers',
          '@google-cloud/spanner',
          '@mongodb-js/zstd',
          '@nestjs/core',
          '@nestjs/microservices',
          '@nestjs/microservices/microservices-module',
          '@nestjs/platform-express',
          '@nestjs/websockets',
          '@nestjs/websockets/socket-module',
          '@sap/hana-client',
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
        ];
        if (!lazyImports.includes(resource)) return false;
        try {
          require.resolve(resource, { paths: [process.cwd()] });
          return false;
        } catch {
          return true;
        }
      },
    }),
    isDev && new rspack.HotModuleReplacementPlugin(),
    isDev && new RunScriptWebpackPlugin({ name: 'main.cjs', autoRestart: false }),
  ],
};

export default config;
