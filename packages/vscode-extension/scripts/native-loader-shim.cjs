// Staged by scripts/build.js into the packaged extension as
// `dist/eslint-plugin/node_modules/@rslint/native/index.js` — a minimal,
// deterministic replacement for napi-rs's auto-generated loader.
//
// The vsix is per-platform and ships exactly one binary as `./rslint.node`, so
// none of the loader's runtime platform/libc/universal probing or version
// checks are needed (and that probing is a liability in the Electron host — its
// win32 branch keys on `process.config.variables.node_target_type`, and
// Electron embeds Node as a shared library). Re-exporting `parse` explicitly
// lets the worker's ESM `import { parse }` resolve the named export from this
// CommonJS module.
const binding = require('./rslint.node');
module.exports = binding;
module.exports.parse = binding.parse;
