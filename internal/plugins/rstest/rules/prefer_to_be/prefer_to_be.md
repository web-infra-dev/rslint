# prefer-to-be

## Rule Details

This rule asks equality assertions to use the matcher that most directly states what is being checked. Primitive literals use `toBe()`, while `null`, `undefined` and `NaN` use their dedicated matchers. Negated undefined checks are normalized to the positive complementary matcher.

Examples of **incorrect** code for this rule:

```ts
expect(status).toEqual('ready');
expect(result).toStrictEqual(1);
expect(value).toBe(null);
expect(value).not.toEqual(undefined);
expect(value).toStrictEqual(NaN);
```

Examples of **correct** code for this rule:

```ts
expect(status).toBe('ready');
expect(result).toBe(1);
expect(value).toBeNull();
expect(value).toBeDefined();
expect(value).toBeNaN();

expect(result).toEqual({ value: 1 });
expect(total).toBeCloseTo(0.3);
```

Fractional numeric literals are left alone. Rstest uses Vitest's assertion implementation, whose public API recommends `toBeCloseTo()` for floating-point comparisons; automatically changing `toEqual(0.3)` to `toBe(0.3)` would preserve exact equality while steering the test toward the wrong matcher.

The rule recognizes Rstest globals, imports from Rstest packages, namespace and `require` bindings, `import.meta.rstest`, `expect.soft()`, `expect.poll()`, and the `expect` supplied by a [test context](https://rstest.rs/api/runtime-api/test-api/test#testcontext). Browser-only `expect.element()` assertions are excluded because their matcher set does not provide these value matchers. Chai-style assertions such as `expect(value).to.equal(1)` are also left alone. Locally declared values named `undefined` or `NaN` do not select the dedicated built-in matchers.

## Autofix

The autofix renames the matcher, removes arguments from dedicated no-argument matchers, and removes `.not` when a negated undefined check is rewritten to its positive complement. The authored accessor style is retained, so `["toEqual"]` becomes `["toBe"]`. If removing arguments would delete a comment, the diagnostic is emitted without a fix.
