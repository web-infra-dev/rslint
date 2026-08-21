# no-alias-methods

## Rule Details

This rule disallows deprecated alias matchers and prefers their canonical matcher names.

Examples of **incorrect** code for this rule:

```typescript
expect(spy).toBeCalled();
expect(spy).lastCalledWith('ready');
expect(error).not['toThrowError']();
```

Examples of **correct** code for this rule:

```typescript
expect(spy).toHaveBeenCalled();
expect(spy).toHaveBeenLastCalledWith('ready');
expect(error).not['toThrow']();
```

## Options

This rule has no options.

## Autofix

This rule autofixes supported accessor forms and preserves the original accessor delimiter:

- `expect(spy).toBeCalled()` becomes `expect(spy).toHaveBeenCalled()`
- `expect(spy)['toBeCalled']()` becomes `expect(spy)['toHaveBeenCalled']()`
- `expect(spy)["toBeCalled"]()` becomes `expect(spy)["toHaveBeenCalled"]()`
- ``expect(spy)[`toBeCalled`]()`` becomes ``expect(spy)[`toHaveBeenCalled`]()```

Terminal computed identifier keys such as `expect(spy)[matcherName]()` are reported only with no prefix or a valid `expect` modifier prefix. Chai-prefixed dynamic keys are ignored, and reported dynamic keys are never fixed.

## Differences from eslint-plugin-vitest

Rslint uses the same matcher mapping as `eslint-plugin-vitest@v1.6.27`, but two user-visible details are different:

- the diagnostic wording says `canonical name of ...`;
- element-access fixes preserve the original quote style or template-literal delimiter instead of rewriting it to a single-quoted string.

## Original Documentation

- [eslint-plugin-vitest: no-alias-methods](https://github.com/vitest-dev/eslint-plugin-vitest/blob/v1.6.27/docs/rules/no-alias-methods.md)
- [Source code](https://github.com/vitest-dev/eslint-plugin-vitest/blob/v1.6.27/src/rules/no-alias-methods.ts)
