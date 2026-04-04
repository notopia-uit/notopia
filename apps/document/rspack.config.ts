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
  node: {
    __dirname: false,
    __filename: false,
  },
  optimization: {
    minimize: false,
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
    nodeExternals({
      importType: isEsm ? 'module' : 'commonjs',
      allowlist: [
        // /^@notopia-uit/,
        // /^@blocknote/,
        // /^@handlewithcare/,
        // /^prosemirror/,
      ],
    }) as any,
    ...(isEsm
      ? [
          (data, callback: any) => {
            const request = data.request;
            if (!request) return callback();
            const bare = request.startsWith('node:')
              ? request.slice(5)
              : request;
            if (builtinModules.includes(bare)) {
              return callback(null, `node:${bare}`);
            }
            callback();
          },
        ]
      : []),
  ],
  plugins: [
    new NxAppRspackPlugin({
      tsConfig: tsConfigFile,
      main: 'apps/document/src/main.ts',
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
          '@nestjs/microservices',
          '@nestjs/websockets',
          'cache-manager',
          'class-validator',
          'class-transformer',
          'amqp-connection-manager',
          'amqplib',
          'mqtt',
          'nats',
          'ioredis',
          'redis',
          'kafkajs',
          'pg-native',
          'canvas',
          // TypeORM optional dependencies
          'better-sqlite3',
          'sqlite3',
          'mysql',
          'mysql2',
          'oracledb',
          'pg-query-stream',
          'spanner',
          'sql.js',
          'typeorm-aurora-data-api-driver',
          // Other optional deps
          'hiredis',
          '@mongodb-js/zstd',
          'kerberos',
          'snappy',
          '@aws-sdk/credential-providers',
          'gcp-metadata',
          // NestJS file validation (has broken exports)
          'file-type',
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
