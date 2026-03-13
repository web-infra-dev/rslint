import type { RslintConfigEntry } from '../define-config.js';

// Aligned with official eslint-plugin-import recommended.
// Rules commented out with "not implemented" are in the official preset but not yet available.
const recommended: RslintConfigEntry = {
  plugins: ['eslint-plugin-import'],
  rules: {
    // errors
    // 'import/no-unresolved': 'error', // not implemented
    // 'import/named': 'error', // not implemented
    // 'import/namespace': 'error', // not implemented
    // 'import/default': 'error', // not implemented
    // 'import/export': 'error', // not implemented
    // warnings
    // 'import/no-named-as-default': 'warn', // not implemented
    // 'import/no-named-as-default-member': 'warn', // not implemented
    // 'import/no-duplicates': 'warn', // not implemented
  },
};

export { recommended };
