# expect-expect

## Rule Details

Ensure every Jest test callback contains at least one assertion. The rule tracks test APIs such as `test`, `it`, `fit`, `xit`, and `xtest` (including chained forms like `it.each` that the Jest integration recognizes) and reports when none of the configured assertion callee patterns appear in the body. Assertions inside a named function declaration or variable function passed as the test callback are attributed to every test that references it, regardless of whether the callback is declared before or after the registration. This guards against tests that run side effects but never verify outcomes.

Skipped [`test.todo` / `it.todo`](https://jestjs.io/docs/api#testtodotitle) bodies are ignored.

### Divergence from `eslint-plugin-jest`

For a named callback whose declaration is asserting, this rule intentionally treats the registration as covered **independent of source order**, whereas upstream `eslint-plugin-jest` only clears a registration if it was already seen when the assertion was walked. Concretely:

```js
function myTest() {
  expect(true).toBeDefined();
}
it('should pass', myTest);
```

Upstream reports `Test has no assertions` here (the assertion is walked before the `it(...)` registration is queued), while this rule does not. Because the function declaration is hoisted, this snippet is runtime-equivalent to the call-first form `it('should pass', myTest); function myTest() { ... }`, which both implementations already accept — so treating the two forms identically is the correct behavior and removes an order-dependent false positive.

Examples of **incorrect** code for this rule:

```js
it('should be a test', () => {
  console.log('no assertion');
});
test('should assert something', () => {});
```

Examples of **correct** code for this rule:

```js
it('should be a test', () => {
  expect(true).toBeDefined();
});
it('should work with callbacks/async', () => {
  somePromise().then(res => expect(res).toBe('passed'));
});
```

### Options

```ts
interface Options {
  assertFunctionNames?: string[];
  additionalTestBlockFunctions?: string[];
}
```

- **`assertFunctionNames`** (default `["expect"]`): callee chains that count as assertions. Patterns follow [eslint-plugin-jest](https://github.com/jest-community/eslint-plugin-jest/blob/main/docs/rules/expect-expect.md): `*` matches a dot-separated segment; `**` matches zero or more segments. The pattern is matched case-insensitively against the full chain (for example `request.**.expect` for [SuperTest](https://www.npmjs.com/package/supertest) `.expect`). Special regex characters in names may need escaping when mirroring ESLint behavior.
- **`additionalTestBlockFunctions`**: extra global function names treated like `test`/`it` wrappers (for example helpers from [`jest-theories`](https://www.npmjs.com/package/jest-theories)) so their callbacks are also required to contain an assertion.

For more option examples and edge cases, see the upstream rule documentation linked below.

## Original Documentation

- [jest/expect-expect](https://github.com/jest-community/eslint-plugin-jest/blob/main/docs/rules/expect-expect.md)
