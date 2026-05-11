import pluginX from './plugin-x.mjs';

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
      px: pluginX,
    },
    rules: {
      'px/no-foo': 'error',
    },
  },
];
