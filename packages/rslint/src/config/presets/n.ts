import type { RslintConfigEntry } from '../define-config.js';

// Aligned with official eslint-plugin-n@18.2.1 recommended.
// Rules commented out with "not implemented" are in the official preset but not yet available.
const recommended: RslintConfigEntry = {
  plugins: ['n'],
  rules: {
    // 'n/hashbang': 'error', // not implemented
    'n/no-deprecated-api': 'error',
    // 'n/no-exports-assign': 'error', // not implemented
    // 'n/no-extraneous-import': 'error', // not implemented
    // 'n/no-extraneous-require': 'error', // not implemented
    // 'n/no-missing-import': 'error', // not implemented
    // 'n/no-missing-require': 'error', // not implemented
    // 'n/no-process-exit': 'error', // not implemented
    // 'n/no-unpublished-bin': 'error', // not implemented
    // 'n/no-unpublished-import': 'error', // not implemented
    // 'n/no-unpublished-require': 'error', // not implemented
    // 'n/no-unsupported-features/es-builtins': 'error', // not implemented
    // 'n/no-unsupported-features/es-syntax': 'error', // not implemented
    // 'n/no-unsupported-features/node-builtins': 'error', // not implemented
    // 'n/process-exit-as-throw': 'error', // not implemented
    // 'n/shebang': 'error', // not implemented
  },
};

export { recommended };
