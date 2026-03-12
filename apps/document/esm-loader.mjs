/**
 * ESM Loader Hook for Node.js that works around the @blocknote ESM compatibility issue
 * This allows CommonJS modules with missing "exports" fields to be loaded via require()
 */

export async function resolve(specifier, context, nextResolve) {
  // Use the default resolution for now
  return nextResolve(specifier, context);
}

export async function load(url, context, nextLoad) {
  // If the module is @handlewithcare/prosemirror-inputrules, skip the exports validation
  if (url.includes('@handlewithcare/prosemirror-inputrules')) {
    return nextLoad(url, { ...context, format: 'commonjs' });
  }

  return nextLoad(url, context);
}
