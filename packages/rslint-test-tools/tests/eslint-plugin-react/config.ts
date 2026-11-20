import type { RuleTesterConfig } from '@rslint/rule-tester';

// Preserve the original projects and virtual TSX filename.
const compilerOptions = {
  jsx: 'react',
  target: 'es5',
  module: 'commonjs',
  strict: true,
  esModuleInterop: true,
  lib: ['es2015', 'es2017', 'esnext'],
};

export const fixtureConfig: RuleTesterConfig = {
  plugin: 'react',
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

// Parser metadata shared by the corresponding pinned upstream groups.
export const upstreamModuleOptions: NonNullable<
  RuleTesterConfig['languageOptions']
> = {
  ecmaVersion: 2018,
  sourceType: 'module',
  parserOptions: {
    ecmaFeatures: {
      jsx: true,
    },
  },
};

// Parser metadata shared by the corresponding pinned upstream groups.
export const upstreamScriptOptions: NonNullable<
  RuleTesterConfig['languageOptions']
> = {
  ecmaVersion: 2018,
  sourceType: 'script',
  parserOptions: {
    ecmaFeatures: {
      jsx: true,
    },
  },
};

// Parser metadata shared by the corresponding pinned upstream groups.
export const upstreamES2015Options: NonNullable<
  RuleTesterConfig['languageOptions']
> = {
  ecmaVersion: 2015,
  sourceType: 'module',
  parserOptions: {
    ecmaFeatures: {
      jsx: true,
    },
  },
};
