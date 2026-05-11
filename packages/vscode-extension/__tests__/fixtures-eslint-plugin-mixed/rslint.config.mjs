import fixturePlugin from './plugin.mjs';

export default [
  {
    files: ['**/*.ts'],
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
      // Native syntax-only rule — fires on `debugger;` in both program
      // files and gap files (needs no type info).
      'no-debugger': 'error',
      // ESLint-plugin rule routed through the worker.
      'fx/no-forbidden': 'error',
    },
  },
];
