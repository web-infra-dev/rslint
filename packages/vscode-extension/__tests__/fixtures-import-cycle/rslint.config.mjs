export default [
  {
    files: ['**/*.ts'],
    languageOptions: {
      parserOptions: {
        projectService: false,
        project: ['./tsconfig.json'],
      },
    },
    plugins: ['import'],
    rules: {
      'import/no-cycle': 'error',
      'no-var': 'error',
    },
  },
];
