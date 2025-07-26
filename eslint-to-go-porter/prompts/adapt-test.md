You are adapting a TypeScript ESLint test file to work with the rslint cross-validation test framework.

## Original TypeScript ESLint Test

```typescript
{{TEST_SOURCE}}
```

## Instructions

Convert this test file to work with rslint's cross-validation test framework. You need to:

1. **Update imports**: Change from `@typescript-eslint/rule-tester` to the rslint format:
   ```typescript
   import { RuleTester, getFixturesRootDir } from '../RuleTester.ts';
   ```
   Optional: Import `noFormat` if the test uses unformatted code samples

2. **Remove rule import**: Remove the `import rule from '../../src/rules/...'` line since rslint runs rules by name

3. **Update RuleTester construction**: Use the rslint format:
   ```typescript
   const rootPath = getFixturesRootDir();
   const ruleTester = new RuleTester({
     languageOptions: {
       parserOptions: {
         project: './tsconfig.json',
         tsconfigRootDir: rootPath,
       },
     },
   });
   ```

4. **Update run call**: Remove the rule parameter:
   ```typescript
   ruleTester.run('{{RULE_NAME}}', {
     valid: [...],
     invalid: [...]
   });
   ```

5. **Keep test cases identical**: Preserve all valid and invalid test cases exactly as they are

6. **Test Case Format Notes**:
   - Line and column numbers are 1-based (not 0-based)
   - Message IDs should be camelCase and match the rule's defined message IDs
   - Options are passed as an array: `options: [{ configOption: 'value' }]`
   - For invalid cases with fixes, include `output: 'fixed code'` or `output: null` if no fix
   - Suggestions array format: `suggestions: [{ messageId: 'suggestionId', output: 'fixed code' }]`

7. **IMPORTANT**: Do NOT attempt to run, compile, or execute the test file
8. **IMPORTANT**: Do NOT create any temporary files or debug output

## Output

Create ONLY the adapted test file at: `packages/rslint-test-tools/tests/typescript-eslint/rules/{{RULE_NAME}}.test.ts`

Provide ONLY the adapted TypeScript code without explanations. Do not create any other files.