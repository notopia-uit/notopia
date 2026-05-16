"use strict";

// Workaround for @protobufjs/inquire in rspack/webpack bundles.
// The original uses eval("require") to escape the bundler, but rspack replaces
// that with an empty context module, so inquire("fs") returns null at runtime,
// crashing Root.loadSync with "Cannot read properties of null (reading 'readFileSync')".
// Upstream fix (webpackIgnore magic comment) is in protobufjs PR#2226, not yet released.
const fs = require("fs");
const path = require("path");
const buffer = require("buffer");

let long = null;
try { long = require("long"); } catch (e) {}

const _modules = { fs, path, long, buffer };

module.exports = function inquire(moduleName) {
  const mod = _modules[moduleName];
  if (mod && (mod.length || Object.keys(mod).length)) return mod;
  return null;
};
