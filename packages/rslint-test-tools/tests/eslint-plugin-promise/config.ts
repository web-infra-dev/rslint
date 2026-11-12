import type { RuleTesterConfig } from '@rslint/rule-tester';

// Preserve the suite's TypeScript project and virtual TSX filename.
export const config: RuleTesterConfig = {
  plugin: 'promise',
  filename: 'src/virtual.tsx',
  fixtureFiles: {
    'tsconfig.json': JSON.stringify({
      compilerOptions: {
        jsx: 'react',
        target: 'es5',
        module: 'commonjs',
        strict: true,
        esModuleInterop: true,
        lib: ['es2015', 'es2017', 'esnext'],
      },
      include: ['src/**/*'],
    }),
  },
  languageOptions: { parserOptions: { project: ['./tsconfig.json'] } },
};
