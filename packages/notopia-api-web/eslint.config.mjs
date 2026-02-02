import nx from "@nx/eslint-plugin";
import baseConfig from "../../eslint.config.mjs";
import { defineConfig } from "eslint/config";

export default defineConfig(...baseConfig, ...nx.configs["flat/react"], {
  files: ["**/*.ts", "**/*.tsx", "**/*.js", "**/*.jsx"],
});
