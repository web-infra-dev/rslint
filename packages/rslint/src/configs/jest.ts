import type { RslintConfigEntry } from '../define-config.js';

// Aligned with official eslint-plugin-jest@29.x recommended.
// Rules commented out with "not implemented" are in the official preset but not yet available.
const recommended: RslintConfigEntry = {
  plugins: ['jest'],
  rules: {
    // 'jest/expect-expect': 'warn', // not implemented
    'jest/no-alias-methods': 'error',
    // 'jest/no-commented-out-tests': 'warn', // not implemented
    // 'jest/no-conditional-expect': 'error', // not implemented
    // 'jest/no-deprecated-functions': 'error', // not implemented
    'jest/no-disabled-tests': 'warn',
    // 'jest/no-done-callback': 'error', // not implemented
    // 'jest/no-export': 'error', // not implemented
    'jest/no-focused-tests': 'error',
    // 'jest/no-identical-title': 'error', // not implemented
    // 'jest/no-interpolation-in-snapshots': 'error', // not implemented
    // 'jest/no-jasmine-globals': 'error', // not implemented
    // 'jest/no-mocks-import': 'error', // not implemented
    // 'jest/no-standalone-expect': 'error', // not implemented
    'jest/no-test-prefixes': 'error',
    'jest/valid-describe-callback': 'error',
    // 'jest/valid-expect': 'error', // not implemented
    // 'jest/valid-expect-in-promise': 'error', // not implemented
    // 'jest/valid-title': 'error', // not implemented
  },
};

export { recommended };
