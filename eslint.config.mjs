import nx from "@nx/eslint-plugin";

export default [
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
];
