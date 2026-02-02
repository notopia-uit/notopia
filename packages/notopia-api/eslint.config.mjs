import baseConfig from "../../eslint.config.mjs";
import { defineConfig } from "eslint/config";

export default defineConfig(...baseConfig, {
  linterOptions: {
    reportUnusedDisableDirectives: "off",
  },
});
