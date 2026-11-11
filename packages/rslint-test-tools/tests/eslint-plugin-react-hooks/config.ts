import type { RuleTesterConfig } from '@rslint/rule-tester';

// Cases lint as a TSX file in a TypeScript project.
export const config: RuleTesterConfig = {
  plugin: 'react-hooks',
  filename: 'src/virtual.tsx',
  fixtureFiles: {
    'tsconfig.json': JSON.stringify({
      compilerOptions: {
        jsx: 'react',
        target: 'esnext',
        module: 'commonjs',
        strict: false,
        esModuleInterop: true,
        lib: ['es2015', 'es2017', 'esnext'],
      },
      include: ['src/**/*'],
    }),
  },
  languageOptions: { parserOptions: { project: ['./tsconfig.json'] } },
};
