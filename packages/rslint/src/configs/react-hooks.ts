import type { RslintConfigEntry } from '../define-config.js';

// Aligned with official eslint-plugin-react-hooks@7.x recommended.
// Rules commented out with "not implemented" are in the official preset but not yet available.
const recommended: RslintConfigEntry = {
  plugins: ['react-hooks'],
  rules: {
    'react-hooks/rules-of-hooks': 'error',
    'react-hooks/exhaustive-deps': 'warn',
    // 'react-hooks/config': 'error', // not implemented
    // 'react-hooks/error-boundaries': 'error', // not implemented
    // 'react-hooks/gating': 'error', // not implemented
    // 'react-hooks/globals': 'error', // not implemented
    // 'react-hooks/immutability': 'error', // not implemented
    // 'react-hooks/incompatible-library': 'warn', // not implemented
    // 'react-hooks/preserve-manual-memoization': 'error', // not implemented
    // 'react-hooks/purity': 'error', // not implemented
    // 'react-hooks/refs': 'error', // not implemented
    // 'react-hooks/set-state-in-effect': 'error', // not implemented
    // 'react-hooks/set-state-in-render': 'error', // not implemented
    // 'react-hooks/static-components': 'error', // not implemented
    // 'react-hooks/unsupported-syntax': 'warn', // not implemented
    // 'react-hooks/use-memo': 'error', // not implemented
  },
};

export { recommended };
