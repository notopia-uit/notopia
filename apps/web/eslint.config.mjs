import { defineConfig } from "eslint/config";
import baseConfig from "../../eslint.config.mjs";
import eslintPluginBetterTailwindcss from "eslint-plugin-better-tailwindcss";
import nextEslintPluginNext from "@next/eslint-plugin-next";
import nx from "@nx/eslint-plugin";
import reactHooks from "eslint-plugin-react-hooks";

export default defineConfig(
  ...baseConfig,
  ...reactHooks.configs.flat.recommended,
  {
    plugins: {
      "@next/next": nextEslintPluginNext,
    },
  },
  ...nx.configs["flat/react-typescript"],
  {
    extends: [eslintPluginBetterTailwindcss.configs.recommended],
    settings: {
      "better-tailwindcss": {
        entryPoint: "./src/app/global.css",
      },
    },
  },
  {
    ignores: [".next/**/*", "**/out-tsc"],
  },
);
