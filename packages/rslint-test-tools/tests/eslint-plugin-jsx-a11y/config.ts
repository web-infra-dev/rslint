import type { RuleTesterConfig } from '@rslint/rule-tester';

const compilerOptions = {
  jsx: 'react',
  target: 'es5',
  module: 'commonjs',
  strict: true,
  esModuleInterop: true,
  lib: ['es2015', 'es2017', 'esnext'],
};

// Preserve the original TypeScript projects, JSX mode and virtual filename.
export const config: RuleTesterConfig = {
  plugin: 'jsx-a11y',
  filename: 'src/virtual.tsx',
  fixtureFiles: {
    'tsconfig.files.json': JSON.stringify({
      compilerOptions: { ...compilerOptions, allowJs: true },
      include: ['files/**/*'],
    }),
    'tsconfig.virtual.json': JSON.stringify({
      compilerOptions,
      include: ['src/virtual.tsx', 'src/virtual.ts'],
    }),
  },
  languageOptions: {
    parserOptions: {
      projectService: false,
      project: ['./tsconfig.files.json', './tsconfig.virtual.json'],
    },
  },
};
