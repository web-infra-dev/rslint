// CJS-style `module.exports = X` — TypeScript treats this as `export = X`,
// which under esModuleInterop becomes the synthesized default.
module.exports = function cjsDefault() {
  return 'cjs';
};
