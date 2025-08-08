import { lint } from './packages/rslint/dist/index.js';

const result = await lint({
  config: '/Users/bytedance/dev/rslint/packages/rslint-test-tools/tests/typescript-eslint/fixtures/rslint.json',
  workingDirectory: '/Users/bytedance/dev/rslint/packages/rslint-test-tools/tests/typescript-eslint/fixtures',
  fileContents: {
    'src/virtual.ts': `// no accessibility === public
interface Foo {
  [Z: string]: any;
  A: string;
  B: string;
  C: string;
  D: string;
  E: string;
  F: string;
  new ();
  G();
  H();
  I();
  J();
  K();
  L();
}`
  },
  ruleOptions: {
    '@typescript-eslint/member-ordering': []
  }
});

console.log('Result:', JSON.stringify(result, null, 2));