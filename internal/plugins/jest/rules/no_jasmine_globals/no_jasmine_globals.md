# no-jasmine-globals

## Rule Details

Jest supports running without Jasmine globals. This rule disallows using Jasmine-specific globals and APIs in tests, and requires the Jest equivalents instead.

Examples of **incorrect** code for this rule:

```typescript
spyOn(obj, 'method');
fail();
pending();
jasmine.addMatchers(matchers);
jasmine.createSpy();
jasmine.any(Number);
jasmine.DEFAULT_TIMEOUT_INTERVAL = 5000;
```

Examples of **correct** code for this rule:

```typescript
jest.spyOn(obj, 'method');
throw new Error('failed');
test.skip('skipped for now', () => {});
expect.extend(matchers);
jest.fn();
expect.any(Number);
jest.setTimeout(5000);
```

## Original Documentation

- [jest/no-jasmine-globals](https://github.com/jest-community/eslint-plugin-jest/blob/main/docs/rules/no-jasmine-globals.md)
