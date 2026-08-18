import type { RslintConfigEntry } from '../define-config.js';

const recommended: RslintConfigEntry = {
  plugins: ['rstest'],
  rules: {
    'rstest/expect-expect': 'warn',
    'rstest/no-commented-out-tests': 'warn',
    'rstest/no-conditional-expect': 'error',
    'rstest/no-disabled-tests': 'warn',
    'rstest/no-focused-tests': 'error',
    'rstest/no-identical-title': 'error',
    'rstest/no-import-node-test': 'error',
    'rstest/no-interpolation-in-snapshots': 'error',
    'rstest/no-mocks-import': 'error',
    'rstest/no-standalone-expect': 'error',
    'rstest/require-local-test-context-for-concurrent-snapshots': 'error',
    'rstest/valid-expect': 'error',
    'rstest/valid-expect-in-promise': 'error',
    'rstest/valid-title': 'error',
  },
};

export { recommended };
