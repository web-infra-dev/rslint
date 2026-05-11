import fixturePlugin from './plugin.mjs';

export default [
  {
    files: ['src/**/*.ts'],
    languageOptions: {
      parserOptions: {
        projectService: false,
        project: ['./tsconfig.json'],
      },
    },
    eslintPlugins: {
      fx: fixturePlugin,
    },
    rules: {
      'fx/no-forbidden': 'error',
      'fx/rename-banned': 'error',
    },
  },
];
