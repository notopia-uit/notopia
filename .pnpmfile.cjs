/**
 * pnpm hook file that patches @handlewithcare/prosemirror-inputrules
 * to properly define its exports for ESM compatibility
 */
function afterAllResolved(lockfile) {
  Object.keys(lockfile.packages).forEach((pkgPath) => {
    if (pkgPath.includes('@handlewithcare/prosemirror-inputrules')) {
      const pkg = lockfile.packages[pkgPath];

      // Ensure the package has proper exports defined
      if (!pkg.exports) {
        pkg.exports = {
          '.': {
            import: './dist/index.mjs',
            require: './dist/index.cjs',
            default: './dist/index.mjs',
          },
        };
      }

      // Mark it as having proper ESM support
      if (!pkg.main) {
        pkg.main = './dist/index.cjs';
      }
    }
  });

  return lockfile;
}

module.exports = {
  hooks: {
    afterAllResolved,
  },
};
