// Root config has NO eslintPlugins. Each nested package brings its
// own. Files matched by this root entry that aren't inside any
// nested package's `files` glob fall back here — but in this fixture
// the only source lives under packages/*/src, so root effectively
// just exists to anchor tsconfig discovery.
export default [
  {
    files: ['packages/*/src/**/*.ts'],
    languageOptions: {
      parserOptions: {
        projectService: false,
        project: ['./tsconfig.json'],
      },
    },
  },
];
