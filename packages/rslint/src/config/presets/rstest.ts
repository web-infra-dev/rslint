import type { RslintConfigEntry } from '../define-config.js';

const recommended: RslintConfigEntry = {
  plugins: ['rstest'],
  rules: {
    'rstest/no-focused-tests': 'error',
    'rstest/no-mocks-import': 'error',
  },
};

export { recommended };
