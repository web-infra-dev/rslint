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

6. **IMPORTANT**: Do NOT attempt to run, compile, or execute the test file

## Output

Create the adapted test file at: `packages/rslint-test-tools/tests/typescript-eslint/rules/{{RULE_NAME}}.test.ts`

Provide ONLY the adapted TypeScript code without explanations.