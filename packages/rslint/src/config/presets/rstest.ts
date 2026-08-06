import type { RslintConfigEntry } from '../define-config.js';

const recommended: RslintConfigEntry = {
  plugins: ['rstest'],
  rules: {
    'rstest/no-commented-out-tests': 'warn',
    'rstest/no-conditional-expect': 'error',
    'rstest/no-disabled-tests': 'warn',
    'rstest/no-identical-title': 'error',
    'rstest/no-mocks-import': 'error',
    'rstest/valid-title': 'error',
  },
};

export { recommended };
