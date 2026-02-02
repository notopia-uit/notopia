import nextEslintPluginNext from "@next/eslint-plugin-next";
import nx from "@nx/eslint-plugin";
import baseConfig from "../../eslint.config.mjs";
import { defineConfig } from "eslint/config";

export default defineConfig(
  { plugins: { "@next/next": nextEslintPluginNext } },
  ...baseConfig,
  ...nx.configs["flat/react-typescript"],
  {
    ignores: [".next/**/*", "**/out-tsc"],
  },
);
