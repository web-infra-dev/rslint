# prefer-called-times

## Rule Details

Prefer `expect(fn).toHaveBeenCalledTimes(1)` over
`expect(fn).toHaveBeenCalledOnce()`. Both assert the same thing, and spelling
the count out keeps a file's call-count assertions in one form, so changing an
expected count is an edit to the number rather than a switch to a different
matcher.

Examples of **incorrect** code:

```ts
expect(fn).toHaveBeenCalledOnce();
expect(fn).not.toHaveBeenCalledOnce();
expect(fn).resolves.toHaveBeenCalledOnce();
expect.soft(fn).toHaveBeenCalledOnce();
```

Examples of **correct** code:

```ts
expect(fn).toHaveBeenCalledTimes(1);
expect(fn).not.toHaveBeenCalledTimes(1);
expect(fn).toHaveBeenCalledTimes(2);
```

## Options

This rule has no options.

## Recognized assertions

The matcher has to be called. `expect(fn).toHaveBeenCalledOnce` asserts
nothing, and rewriting it would turn a no-op into a live assertion, so it is
left alone. So is `expect.toHaveBeenCalledOnce()`, which never received a value
to count calls on.

Chai's `calledOnce` is not covered. It is a property rather than a matcher
call, and the assertion that takes a count in that style is `callCount(1)`, so
`expect(fn).to.have.been.calledOnce` is not reported.

A matcher named by a computed key, `expect(fn)[matcherName]()`, is not
reported: which assertion runs is only known at runtime.

Parentheses around the assertion are transparent, including around an optional
chain: `(expect(fn)?.toHaveBeenCalledOnce)()` is reported and fixed.

Every expect source is covered: globals, `@rstest/core` named imports and
renamed imports, `require('@rstest/core')` destructuring, namespace imports,
whole-module `require`, `import.meta.rstest`, `@rstest/playwright`, and the
`expect` a test callback receives through its context (`({ expect })` and
`ctx.expect`). An `expect` from another assertion library, and a local variable
that shadows `expect`, are not reported.

## Fix

The fix renames the matcher and gives it the count `1`, leaving the rest of the
assertion exactly as written. The expect root as spelled at the call site,
`expect.soft`, the optional second `message` argument of
`expect(actual, message)`, the modifier chain, the accessor's own quoting, and
the surrounding line breaks and comments all survive:

```ts
expect.soft(fn, 'called once')['toHaveBeenCalledOnce']();
// becomes
expect.soft(fn, 'called once')['toHaveBeenCalledTimes'](1);
```

An assertion split across lines, or written with a comment between the matcher
and its parentheses, is fixed as written. A comment already sitting between the
parentheses is kept, with the `1` inserted before it.

The count goes into the matcher's own argument list, never into a call that
merely encloses the assertion, so `expect(fn).toHaveBeenCalledOnce()()` becomes
`expect(fn).toHaveBeenCalledTimes(1)()`.

## Conflicting rules

A rule enforcing the opposite preference — rewriting
`toHaveBeenCalledTimes(1)` into `toHaveBeenCalledOnce()` — must not be enabled
alongside this one, or the two fixes will rewrite each other. Choose the form
the codebase should use and enable only the rule for it.
