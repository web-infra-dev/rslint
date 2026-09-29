import type { GlobalsConfig, RslintConfigEntry } from '../define-config.js';
import { globals } from '../globals/index.js';

// Aligned with official eslint-plugin-jest@29.x flat presets.
// Rules commented out with "not implemented" are in the official preset but not yet available.
const recommended: RslintConfigEntry = {
  plugins: ['jest'],
  languageOptions: {
    get globals(): GlobalsConfig {
      return globals.jest;
    },
  },
  rules: {
    'jest/expect-expect': 'warn',
    'jest/no-alias-methods': 'error',
    'jest/no-commented-out-tests': 'warn',
    'jest/no-conditional-expect': 'error',
    'jest/no-deprecated-functions': 'error',
    'jest/no-disabled-tests': 'warn',
    'jest/no-done-callback': 'error',
    'jest/no-export': 'error',
    'jest/no-focused-tests': 'error',
    'jest/no-identical-title': 'error',
    'jest/no-interpolation-in-snapshots': 'error',
    'jest/no-jasmine-globals': 'error',
    'jest/no-mocks-import': 'error',
    'jest/no-standalone-expect': 'error',
    'jest/no-test-prefixes': 'error',
    'jest/valid-describe-callback': 'error',
    'jest/valid-expect': 'error',
    'jest/valid-expect-in-promise': 'error',
    'jest/valid-title': 'error',
  },
};

const style: RslintConfigEntry = {
  plugins: ['jest'],
  languageOptions: {
    get globals(): GlobalsConfig {
      return globals.jest;
    },
  },
  rules: {
    'jest/prefer-to-be': 'error',
    'jest/prefer-to-contain': 'error',
    'jest/prefer-to-have-length': 'error',
  },
};

const allRuleNames = [
  'jest/consistent-test-it',
  'jest/expect-expect',
  'jest/max-expects',
  'jest/max-nested-describe',
  'jest/no-alias-methods',
  'jest/no-commented-out-tests',
  'jest/no-conditional-expect',
  'jest/no-conditional-in-test',
  'jest/no-confusing-set-timeout',
  'jest/no-deprecated-functions',
  'jest/no-disabled-tests',
  'jest/no-done-callback',
  'jest/no-duplicate-hooks',
  'jest/no-error-equal',
  'jest/no-export',
  'jest/no-focused-tests',
  'jest/no-hooks',
  'jest/no-identical-title',
  'jest/no-interpolation-in-snapshots',
  'jest/no-jasmine-globals',
  // 'jest/no-large-snapshots', // not implemented
  'jest/no-mocks-import',
  'jest/no-restricted-jest-methods',
  'jest/no-restricted-matchers',
  'jest/no-standalone-expect',
  'jest/no-test-prefixes',
  'jest/no-test-return-statement',
  'jest/no-unnecessary-assertion',
  'jest/no-unneeded-async-expect-function',
  'jest/no-untyped-mock-factory',
  'jest/padding-around-after-all-blocks',
  'jest/padding-around-after-each-blocks',
  'jest/padding-around-all',
  'jest/padding-around-before-all-blocks',
  'jest/padding-around-before-each-blocks',
  'jest/padding-around-describe-blocks',
  'jest/padding-around-expect-groups',
  'jest/padding-around-test-blocks',
  'jest/prefer-called-with',
  'jest/prefer-comparison-matcher',
  'jest/prefer-each',
  'jest/prefer-ending-with-an-expect',
  'jest/prefer-equality-matcher',
  'jest/prefer-expect-assertions',
  'jest/prefer-expect-resolves',
  'jest/prefer-hooks-in-order',
  'jest/prefer-hooks-on-top',
  'jest/prefer-importing-jest-globals',
  'jest/prefer-jest-mocked',
  'jest/prefer-lowercase-title',
  'jest/prefer-mock-promise-shorthand',
  'jest/prefer-mock-return-shorthand',
  'jest/prefer-snapshot-hint',
  'jest/prefer-spy-on',
  'jest/prefer-strict-equal',
  'jest/prefer-to-be',
  'jest/prefer-to-contain',
  'jest/prefer-to-have-been-called',
  'jest/prefer-to-have-been-called-times',
  'jest/prefer-to-have-length',
  'jest/prefer-todo',
  'jest/require-hook',
  'jest/require-to-throw-message',
  'jest/require-top-level-describe',
  'jest/unbound-method',
  'jest/valid-describe-callback',
  'jest/valid-expect',
  'jest/valid-expect-in-promise',
  'jest/valid-expect-with-promise',
  'jest/valid-mock-module-path',
  'jest/valid-title',
] as const;

const all: RslintConfigEntry = {
  plugins: ['jest'],
  languageOptions: {
    get globals(): GlobalsConfig {
      return globals.jest;
    },
  },
  rules: Object.fromEntries(allRuleNames.map((name) => [name, 'error'])),
};

export { all, recommended, style };
