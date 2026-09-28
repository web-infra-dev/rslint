import type { RslintConfigEntry } from '../define-config.js';

// Based on official eslint-plugin-import recommended.
const recommended: RslintConfigEntry = {
  plugins: ['eslint-plugin-import'],
  rules: {
    // errors
    'import/no-unresolved': 'error',
    'import/named': 'error',
    'import/namespace': 'error',
    'import/default': 'error',
    'import/export': 'error',
    // warnings
    'import/no-named-as-default': 'warn',
    'import/no-named-as-default-member': 'warn',
    'import/no-duplicates': ['warn', { considerQueryString: true }],
  },
};

export { recommended };
