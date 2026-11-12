import type { RuleTesterConfig } from '@rslint/rule-tester';

export const config: RuleTesterConfig = {
  filename: 'src/virtual.ts',
  fixtureFiles: {
    'tsconfig.json':
      '{"compilerOptions":{"jsx":"preserve","target":"es5","module":"commonjs","strict":true,"esModuleInterop":true,"lib":["es2015","es2017","esnext"],"experimentalDecorators":true},"include":["src/virtual.tsx","src/virtual.ts"]}',
  },
  languageOptions: { parserOptions: { project: ['./tsconfig.json'] } },
};
