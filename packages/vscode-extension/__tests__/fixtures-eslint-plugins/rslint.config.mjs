import local from './local-plugin.mjs';

// Mounts a community-style plugin via `plugins` (reverse-dispatched to
// the Node worker by the LSP server) alongside a native rule (`no-console`),
// so the suite can assert plugin + native diagnostics are merged + published.
export default [
  {
    files: ['**/*.ts'],
    plugins: { local },
    rules: {
      'local/no-null': 'error',
      'local/prefer-array-some': 'error',
      'no-console': 'error',
    },
  },
];
