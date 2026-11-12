import type { RuleTesterConfig } from '@rslint/rule-tester';

export const config: RuleTesterConfig = {
  plugin: 'unicorn',
  filename: 'src/virtual.js',
  fixtureFiles: {
    'tsconfig.json':
      '{\n  "compilerOptions": {\n    "jsx": "react",\n    "target": "es5",\n    "module": "commonjs",\n    "strict": true,\n    "esModuleInterop": true,\n    "lib": ["es2015", "es2017", "esnext"],\n    "allowJs": true,\n    "rootDir": "."\n  },\n  "include": ["**/*"]\n}\n',
  },
  languageOptions: {
    parserOptions: {
      project: ['./tsconfig.json'],
    },
  },
};
