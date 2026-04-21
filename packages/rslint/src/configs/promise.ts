import type { RslintConfigEntry } from '../define-config.js';

// Aligned with official eslint-plugin-promise@7.x recommended.
// Rules commented out with "not implemented" are in the official preset but not yet available.
const recommended: RslintConfigEntry = {
  plugins: ['promise'],
  rules: {
    // 'promise/always-return': 'error', // not implemented
    // 'promise/no-return-wrap': 'error', // not implemented
    'promise/param-names': 'error',
    // 'promise/catch-or-return': 'error', // not implemented
    // 'promise/no-native': 'off', // not implemented
    // 'promise/no-nesting': 'warn', // not implemented
    // 'promise/no-promise-in-callback': 'warn', // not implemented
    // 'promise/no-callback-in-promise': 'warn', // not implemented
    // 'promise/avoid-new': 'off', // not implemented
    // 'promise/no-new-statics': 'error', // not implemented
    // 'promise/no-return-in-finally': 'warn', // not implemented
    // 'promise/valid-params': 'warn', // not implemented
  },
};

export { recommended };
