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
- ``expect(spy)[`toBeCalled`]()`` becomes ``expect(spy)[`toHaveBeenCalled`]()``

A computed key whose name is only known at runtime, such as `expect(spy)[matcherName]()`, is never reported.

## Differences from eslint-plugin-vitest

Rslint uses the same matcher mapping as `eslint-plugin-vitest@v1.6.27`, but four user-visible details are different:

- the diagnostic wording says `canonical name of ...`;
- element-access fixes preserve the original quote style or template-literal delimiter instead of rewriting it to a single-quoted string;
- computed identifier keys are never reported. Upstream treats `expect(spy)[matcherName]()` as a supported accessor, matches the *variable's own name* against the alias table, and autofixes it by rewriting the variable reference. That answer does not follow from the code: the variable's value is what names the matcher, so `const toBeCalled = 'toHaveBeenCalled'` is reported and rewritten even though it already calls the canonical matcher. This rule stays silent instead;
- template-literal keys are compared by their cooked value, so ``expect(spy)[`toBeCall\u0065d`]()`` is reported and fixed. Upstream compares `quasis[0].value.raw` and reports nothing. The cooked value is the property name the runtime actually looks up, so this is the more accurate answer, but it does turn upstream-valid code into an error. String-literal keys agree in both implementations, which also use the cooked value.
