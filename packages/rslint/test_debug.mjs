import { lint } from './lib/index.js';

const result = await lint({
  config: [
    {
      language: 'typescript',
      files: ['src/**/*.ts'],
      rules: {
        'consistent-indexed-object-style': 'error'
      }
    }
  ],
  fileContents: {
    'src/test.ts': `interface Foo {
  [key: string]: Foo;
}

interface FooUnion {
  [key: string]: Foo | string;
}

interface Simple {
  [key: string]: string;
}`
  }
});

console.log('Result:', JSON.stringify(result, null, 2));