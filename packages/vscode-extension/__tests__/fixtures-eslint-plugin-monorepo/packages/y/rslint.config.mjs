import pluginY from './plugin-y.mjs';

export default [
  {
    files: ['src/**/*.ts'],
    languageOptions: {
      parserOptions: {
        projectService: false,
        project: ['../../tsconfig.json'],
      },
    },
    eslintPlugins: {
      py: pluginY,
    },
    rules: {
      'py/no-bar': 'error',
    },
  },
];
