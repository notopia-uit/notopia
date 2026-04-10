import { NxAppRspackPlugin } from '@nx/rspack/app-plugin';
import type { Configuration } from '@rspack/cli';
import rspack from '@rspack/core';
import { builtinModules } from 'module';
import { join } from 'path';
import nodeExternals from 'webpack-node-externals';

const isEsm = false;
const tsConfigFile = join(__dirname, 'tsconfig.app.json');

const config: Configuration = {
  target: 'node',
  output: {
    module: isEsm,
    path: join(__dirname, './dist'),
    chunkFormat: isEsm ? 'module' : 'array-push',
    filename: '[name].js',
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
              jsc: {
                parser: {
                  syntax: 'typescript',
                  decorators: true,
                },
                transform: {
                  legacyDecorator: true,
                  decoratorMetadata: true,
                },
                target: 'es2021',
              },
              sourceMaps: process.env['NODE_ENV'] !== 'production',
            },
          },
        ],
        exclude: /node_modules/,
        type: isEsm
          ? ('javascript/esm' as const)
          : ('javascript/auto' as const),
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
    // eslint-disable-next-line @typescript-eslint/no-unsafe-call
    nodeExternals({
      importType: isEsm ? 'module' : 'commonjs',
    }) as unknown as Configuration['externals'],
    ...(isEsm
      ? [
          ((data: any, callback: any) => {
            // eslint-disable-next-line @typescript-eslint/no-unsafe-assignment, @typescript-eslint/no-unsafe-member-access
            const request: string | undefined = data.request;
            // eslint-disable-next-line @typescript-eslint/no-unsafe-call, @typescript-eslint/no-unsafe-return
            if (!request) return callback(null);
            const bare = request.startsWith('node:')
              ? request.slice(5)
              : request;
            if (builtinModules.includes(bare)) {
              // eslint-disable-next-line @typescript-eslint/no-unsafe-call, @typescript-eslint/no-unsafe-return
              return callback(null, `node:${bare}`);
            }
            // eslint-disable-next-line @typescript-eslint/no-unsafe-call
            callback(null);
          }) as unknown as Configuration['externals'],
        ]
      : []),
  ] as Configuration['externals'],
  plugins: [
    new NxAppRspackPlugin({
      tsConfig: tsConfigFile,
      main: 'apps/search-worker/src/main.ts',
      sourceMap: process.env['NODE_ENV'] !== 'production',
      optimization: process.env['NODE_ENV'] === 'production',
      externalDependencies: 'none',
    }),
    new rspack.NormalModuleReplacementPlugin(
      /file-type$/,
      require.resolve('./stub.js')
    ),
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
  ],
};

export default config;
