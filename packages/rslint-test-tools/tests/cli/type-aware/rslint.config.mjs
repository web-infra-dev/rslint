import typeAware from './type-aware-plugin.mjs';

// Object-form `plugins: { ta: ... }` mounts a LIVE JS plugin in the worker
// (string-form would be the native-rule whitelist). parserOptions.project makes
// the file type-aware, so Go builds the snapshot the rule queries through.
export default [
  {
    files: ['fixtures/**/*.ts'],
    languageOptions: {
      parserOptions: {
        projectService: false,
        project: ['./tsconfig.json'],
      },
    },
    plugins: { ta: typeAware },
    rules: {
      'ta/no-undefined-union': 'error',
      'ta/report-type-shape': 'warn',
    },
  },
];
