import nestJsRules from '@darraghor/eslint-plugin-nestjs-typed';
import { defineConfig } from 'oxlint';

import baseConfig from '../../oxlint-base.config.mts';

const typeAwareRulesSet = new Set([
  '@darraghor/nestjs-typed/all-properties-have-explicit-defined',
  '@darraghor/nestjs-typed/all-properties-are-whitelisted',
  '@darraghor/nestjs-typed/validated-non-primitive-property-needs-type-decorator',
  '@darraghor/nestjs-typed/validate-nested-of-array-should-set-each',
  '@darraghor/nestjs-typed/api-enum-property-best-practices',
]);

const isEditing = process.env.EDITING === 'true';

const flatRecommendedRules = nestJsRules.configs.flatRecommended[1].rules ?? {};

const oxlintCompatibleRules = isEditing
  ? Object.fromEntries(
      Object.entries(flatRecommendedRules).filter(([ruleName]) => !typeAwareRulesSet.has(ruleName))
    )
  : flatRecommendedRules;

export default defineConfig({
  extends: [baseConfig],
  plugins: ['node', 'vitest'],
  jsPlugins: [
    {
      name: '@darraghor/nestjs-typed',
      specifier: '@darraghor/eslint-plugin-nestjs-typed',
    },
  ],
  ignorePatterns: ['dist/**', 'coverage/**', 'out-tsc/**'],
  env: {
    node: true,
  },
  rules: {
    'oxc/no-barrel-file': 'error',
    ...oxlintCompatibleRules,
  },
});
