# jest/no-mocks-import

## Rule Details

When using `jest.mock`, tests should import from the original module path (for example `./x`), not from `./__mocks__/x`. Importing directly from a `__mocks__` path can leave you with more than one instance of the mocked module that are not the same reference, which is easy to misread and can make assertions fail in surprising ways.

This rule reports `import` declarations and `require()` calls whose module specifier path contains a `__mocks__` segment.

Examples of **incorrect** code for this rule:

```typescript
import thing from './__mocks__/index';
require('./__mocks__/index');
```

Examples of **correct** code for this rule:

```typescript
import thing from 'thing';
require('thing');
```

## Original Documentation

- [jest/no-mocks-import](https://github.com/jest-community/eslint-plugin-jest/blob/main/docs/rules/no-mocks-import.md)
