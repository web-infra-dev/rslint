import type { RslintConfigEntry } from '../define-config.js';

// Aligned with official eslint-plugin-react recommended.
// Rules commented out with "not implemented" are in the official preset but not yet available.
const recommended: RslintConfigEntry = {
  files: ['**/*.jsx', '**/*.tsx'],
  plugins: ['react'],
  rules: {
    // 'react/display-name': 'error', // not implemented
    // 'react/jsx-key': 'error', // not implemented
    // 'react/jsx-no-comment-textnodes': 'error', // not implemented
    // 'react/jsx-no-duplicate-props': 'error', // not implemented
    // 'react/jsx-no-target-blank': 'error', // not implemented
    // 'react/jsx-no-undef': 'error', // not implemented
    'react/jsx-uses-react': 'error',
    'react/jsx-uses-vars': 'error',
    // 'react/no-children-prop': 'error', // not implemented
    // 'react/no-danger-with-children': 'error', // not implemented
    // 'react/no-deprecated': 'error', // not implemented
    // 'react/no-direct-mutation-state': 'error', // not implemented
    // 'react/no-find-dom-node': 'error', // not implemented
    // 'react/no-is-mounted': 'error', // not implemented
    // 'react/no-render-return-value': 'error', // not implemented
    // 'react/no-string-refs': 'error', // not implemented
    // 'react/no-unescaped-entities': 'error', // not implemented
    // 'react/no-unknown-property': 'error', // not implemented
    'react/no-unsafe': 'off',
    // 'react/prop-types': 'error', // not implemented
    'react/react-in-jsx-scope': 'error',
    // 'react/require-render-return': 'error', // not implemented
  },
};

export { recommended };
