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

The rule reports only a matcher that is actually called with exactly one
argument. A broken chain such as `expect(typeof value).toBe`, a spread argument
such as `expect(typeof value).toBe(...types)`, and a Chai property assertion
such as `expect(typeof value).to.be.ok` are all left alone.

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

## Rstest specifics

The assertion is recognised through Rstest's shared expect parser, so every
expect source is covered: globals, `@rstest/core` named imports and renamed
imports, `require('@rstest/core')` destructuring, namespace imports, whole-module
`require`, `import.meta.rstest`, `@rstest/playwright`, and the `expect` a test
callback receives through its context (`({ expect })` and `ctx.expect`). An
`expect` imported from Vitest, Jest, Playwright or Chai, and a local variable
that shadows `expect`, are not reported.

`expect.poll(fn)` and `expect.element(locator)` are excluded: the first takes a
callback and the second a locator, so neither carries the value being type
checked.

## Differences from ESLint

`@vitest/eslint-plugin` rewrites the whole assertion to
`expect(<value>)<modifiers>.toBeTypeOf(<type>)`. On Rstest that rewrite loses
information, so this port edits only the `typeof` operator and the matcher name.
Three shapes come out differently:

- `expect(actual, message)` keeps its message. Rstest's `expect` takes an
  optional second argument, which upstream's whole-call rewrite drops.
- `expect.soft(typeof value).toBe('string')` stays soft. Upstream rewrites it to
  a plain `expect(...)`, which turns a soft assertion into one that aborts the
  test.
- The expect root keeps its spelling. `import.meta.rstest.expect(...)`, a
  renamed import, and a test context's `expect` are all rewritten to a bare
  `expect` by upstream, which may not even be bound in that scope.

The port also narrows one shape:

- A computed identifier key, `expect(typeof value)[matcherName]('string')`, is
  not reported. The matcher is chosen at runtime, so `toBe` there is the name of
  a variable rather than of a matcher, and neither the report nor a rename of
  that variable would be correct.

## Original Documentation

- [@vitest/eslint-plugin: prefer-expect-type-of](https://github.com/vitest-dev/eslint-plugin-vitest/blob/v1.6.27/docs/rules/prefer-expect-type-of.md)
- [Source code](https://github.com/vitest-dev/eslint-plugin-vitest/blob/v1.6.27/src/rules/prefer-expect-type-of.ts)
