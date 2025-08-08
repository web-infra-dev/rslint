import { lint } from './packages/rslint/dist/index.js';

// Test the invalid case that should produce errors using VIRTUAL FILES like the rule tester
const invalidResult = await lint({
  config: '/Users/bytedance/dev/rslint/packages/rslint-test-tools/tests/typescript-eslint/fixtures/rslint.json',
  workingDirectory: '/Users/bytedance/dev/rslint/packages/rslint-test-tools/tests/typescript-eslint/fixtures',
  fileContents: {
    'src/virtual.ts': `class Foo {
  // This should be out of order - method before field
  doSomething() {}
  name: string = '';
}`
  },
  ruleOptions: {
    '@typescript-eslint/member-ordering': []
  },
  languageOptions: {
    parserOptions: {
      project: './tsconfig.virtual.json',
      tsconfigRootDir: '/Users/bytedance/dev/rslint/packages/rslint-test-tools/tests/typescript-eslint/fixtures'
    }
  }
});

console.log('Invalid test result:');
console.log('- Error count:', invalidResult.errorCount);
console.log('- Rule count:', invalidResult.ruleCount);
console.log('- File count:', invalidResult.fileCount);
console.log('- Diagnostics:', invalidResult.diagnostics.length);

if (invalidResult.diagnostics.length > 0) {
  console.log('\nDiagnostics:');
  for (let i = 0; i < invalidResult.diagnostics.length; i++) {
    const d = invalidResult.diagnostics[i];
    console.log(`  ${i+1}. Line ${d.range.start.line}: ${d.message}`);
  }
}