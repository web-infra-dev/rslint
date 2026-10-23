// Upstream's source package keeps version 0.0.0. Rslib needs the compiler's
// major version to select its native declaration backend.
const {
  versionMajorMinor,
} = require('../../typescript-go/packages/typescript/lib/version.cjs');
exports.version = `${versionMajorMinor}.0-source`;
