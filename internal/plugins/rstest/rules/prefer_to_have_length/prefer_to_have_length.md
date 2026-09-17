# prefer-to-have-length

## Rule Details

This rule requires `toHaveLength()` when `toBe()`, `toEqual()` or `toStrictEqual()` compares a value's `.length`. The dedicated matcher states the intended length check directly and produces a more useful failure message.

The rule recognizes Rstest globals, imports and aliases, namespace and `require` bindings, `rstack/test`, `import.meta.rstest`, `expect.soft()`, Playwright's ordinary value assertions, and the `expect` supplied by a [test context](https://rstest.rs/api/runtime-api/test-api/test#testcontext). Dot, string-literal and template-literal accessors are recognized, including negated assertions. It does not require type information.

Optional `.length` access, promise modifiers, `expect.poll()`, browser-only `expect.element()` assertions, dynamic property or matcher names, and Chai-style matchers are left unchanged. When an assertion chain contains multiple matchers or continues into another member access, each `toBe()`, `toEqual()` and `toStrictEqual()` matcher is reported without a fix, unless an earlier Chai matcher such as `property()` has changed the assertion value.

## Incorrect

```ts
expect(files.length).toBe(3);
expect(usernames['length']).toEqual(2);
expect(messages.length).not.toStrictEqual(0);
```

## Correct

```ts
expect(files).toHaveLength(3);
expect(usernames).toHaveLength(2);
expect(messages).not.toHaveLength(0);
```

## Autofix

The autofix removes the `.length` accessor and replaces `toBe` with `toHaveLength` when the receiver is visibly an array, string or function expression and the expected length is a numeric expression other than negative zero. It preserves the Rstest `expect` source, negation, comments outside the removed accessor, and the matcher's dot, quote or template-literal style. Explicit type arguments on `expect()` and `toBe()` are removed because the rewritten calls receive a different value and `toHaveLength` is not generic.

Reading `.length` from another expression can invoke a getter. A receiver whose direct `.length` access cannot be proven inert, a custom assertion message, a non-literal expected value, extra arguments or using the assertion result would move that access across another evaluated expression, so those cases are reported without a fix. Comparisons through `toEqual()` or `toStrictEqual()` are also not fixed because custom equality testers can change their result, while `toHaveLength()` does not use those testers. Negative zero is not fixed because `toBe()` distinguishes it from zero while `toHaveLength()` does not. The fix is also withheld when a removed type argument list contains a comment or an accessor cannot be rewritten safely.
