# prefer-expect-type-of

## Rule Details

Prefer `expect(value).toBeTypeOf(type)` over `expect(typeof value).toBe(type)`.
The dedicated matcher states the intent directly and, when it fails, reports
the value's actual type instead of a plain string comparison.

Examples of **incorrect** code:

```ts
expect(typeof value).toBe('string');
expect(typeof value).toEqual('number');
expect(typeof value).not.toBe('function');
expect.soft(typeof value).toBe('object');
```

Examples of **correct** code:

```ts
expect(value).toBeTypeOf('string');
expect(value).not.toBeTypeOf('function');
expect.soft(value).toBeTypeOf('object');
expect(typeof value === 'string').toBe(true);
```

## Recognized assertions

The rule reports only a matcher that is actually called with exactly one
argument. A broken chain such as `expect(typeof value).toBe`, a spread argument
such as `expect(typeof value).toBe(...types)`, and a Chai property assertion
such as `expect(typeof value).to.be.ok` are all left alone. So is a matcher
named by a computed key, `expect(typeof value)[matcherName]('string')`, because
which assertion runs is only known at runtime.

Every expect source is covered: globals, `@rstest/core` named imports and
renamed imports, `require('@rstest/core')` destructuring, namespace imports,
whole-module `require`, `import.meta.rstest`, `@rstest/playwright`, and the
`expect` a test callback receives through its context (`({ expect })` and
`ctx.expect`). An `expect` from another assertion library, and a local variable
that shadows `expect`, are not reported.

`expect.poll(fn)` and `expect.element(locator)` are excluded: the first takes a
callback and the second a locator, so neither carries the value being type
checked.

## Fix

The fix edits two places and leaves the rest of the call exactly as written: it
removes the `typeof` operator from the assertion's argument, and renames the
matcher accessor to `toBeTypeOf`. Everything else survives — the expect root as
spelled at the call site, `expect.soft`, the optional second `message` argument
of `expect(actual, message)`, the modifier chain, the matcher argument, the
accessor's own quoting, and the surrounding line breaks and comments:

```ts
expect.soft(typeof value, 'should be a string')['toBe']('string');
// becomes
expect.soft(value, 'should be a string')['toBeTypeOf']('string');
```

A comment written between `typeof` and its operand sits inside the removed span
and is removed with it. Every comment outside those two spans is preserved.
