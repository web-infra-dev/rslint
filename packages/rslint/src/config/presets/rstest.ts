import type { RslintConfigEntry } from '../define-config.js';

const recommended: RslintConfigEntry = {
  plugins: ['rstest'],
  rules: {
    'rstest/no-commented-out-tests': 'warn',
    'rstest/no-mocks-import': 'error',
  },
};

export { recommended };
