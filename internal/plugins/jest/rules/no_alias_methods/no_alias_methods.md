# no-alias-methods

## Rule Details

This rule triggers a warning if the alias name, rather than the canonical name, of a method is used.

The rule checks matchers written with dot access or a static bracket key such as `expect(a)['toBeCalled']()`. A matcher named by a variable, such as `expect(a)[matcherName]()`, is ignored because its name is only known at runtime.

Examples of **incorrect** code for this rule:

```javascript
expect(a).toBeCalled();
expect(a).toBeCalledTimes();
expect(a).toBeCalledWith();
expect(a).lastCalledWith();
expect(a).nthCalledWith();
expect(a).toReturn();
expect(a).toReturnTimes();
expect(a).toReturnWith();
expect(a).lastReturnedWith();
expect(a).nthReturnedWith();
expect(a).toThrowError();
```

Examples of **correct** code for this rule:

```javascript
expect(a).toHaveBeenCalled();
expect(a).toHaveBeenCalledTimes();
expect(a).toHaveBeenCalledWith();
expect(a).toHaveBeenLastCalledWith();
expect(a).toHaveBeenNthCalledWith();
expect(a).toHaveReturned();
expect(a).toHaveReturnedTimes();
expect(a).toHaveReturnedWith();
expect(a).toHaveLastReturnedWith();
expect(a).toHaveNthReturnedWith();
expect(a).toThrow();
```

## Original Documentation

- [eslint-plugin-jest: no-alias-methods](https://github.com/jest-community/eslint-plugin-jest/blob/v29.16.0/docs/rules/no-alias-methods.md)
- [Source code](https://github.com/jest-community/eslint-plugin-jest/blob/v29.16.0/src/rules/no-alias-methods.ts)
