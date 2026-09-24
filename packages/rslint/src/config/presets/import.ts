import type { RslintConfigEntry } from '../define-config.js';

// Based on official eslint-plugin-import recommended.
// Rules commented out with "not implemented" are in the official preset but not yet available.
const recommended: RslintConfigEntry = {
  plugins: ['eslint-plugin-import'],
  rules: {
    // errors
    'import/no-unresolved': 'error',
    // 'import/named': 'error', // not implemented
    'import/namespace': 'error',
    'import/default': 'error',
    'import/export': 'error',
    // warnings
    'import/no-named-as-default': 'warn',
    'import/no-duplicates': ['warn', { considerQueryString: true }],
  },
};

export { recommended };
