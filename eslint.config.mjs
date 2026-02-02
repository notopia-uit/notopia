import nx from "@nx/eslint-plugin";
import { defineConfig } from "eslint/config";

export default defineConfig(
  ...nx.configs["flat/base"],
  ...nx.configs["flat/typescript"],
  ...nx.configs["flat/javascript"],
  {
    ignores: [
      "**/dist",
      "**/vite.config.*.timestamp*",
      "**/vitest.config.*.timestamp*",
      "**/test-output",
      "**/out-tsc",
    ],
  },
);
