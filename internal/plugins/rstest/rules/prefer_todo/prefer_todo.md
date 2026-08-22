# prefer-todo

## Rule Details

Prefer `test.todo(...)` over Rstest test registrations that are present in the
file but still have no implementation. The rule reports two cases:

- the callback is missing entirely;
- the callback is an inline empty function body.

Unlike Jest, Rstest accepts both `(title, fn, timeout)` and
`(title, options, fn?)`. This rule recognizes the object-literal options form
and keeps the options object when it rewrites the registration to `.todo`.

Examples of **incorrect** code for this rule:

```ts
test('needs implementation');
test('empty body', () => {});
test('retry later', { timeout: 1000 }, () => {});
test.skip('temporarily empty', () => {});
```

Examples of **correct** code for this rule:

```ts
test.todo('needs implementation');
test('implemented', () => expect(value).toBe(1));
test('indirect callback', callback);
```

## Limitations

Reporting and autofixing are separate decisions, so a registration can be
reported with no fix attached.

### Reported without a fix

A registration the parser has proved to be an Rstest test is always reported,
even when no rewrite delivers a todo test:

- a same-file `const` alias that hides `.skip` in its initializer, such as
  `const skipped = test.skip; skipped('case')`, because the call site has no
  accessor to replace;
- a repeated `.skip` chain, such as `test.skip.skip('case', () => {})`, because
  replacing one accessor leaves the other skip active.

### Not reported at all

The rule stays silent when it cannot prove the call still reaches the Rstest
test API, or when the run mode it would have to preserve is unknown:

- whole-module CommonJS namespace objects, such as
  `const core = require('@rstest/core'); core.test('case')`. Their API
  properties are mutable even when the namespace binding is declared with
  `const`, so `core.test` may no longer be Rstest;
- aliases that hide a modifier or factory other than `.skip`, such as `.fails`,
  `.each(...)`, `.for(...)` or `.extend(...)`, whose run mode the rewrite would
  have to reason about;
- computed member access such as `test[skip]('case', () => {})`. The member is
  not statically known, so the run mode may be one the rule does not report on
  at all, such as `.only` or `.each`.

Sources the rule does follow are Rstest globals, named imports, ES module
namespace imports, `import.meta.rstest` including `const` object destructuring
of it, and same-file `const` aliases whose every hop stays a bare test API.

### Overload recognition

The options overload is recognized only when the second argument is written as
an object literal. Calls such as `test('case', options)` and
`test('case', options, () => {})` are left unchanged because the identifier may
still be the callback.

## Original Documentation

- [eslint-plugin-jest: prefer-todo](https://github.com/jest-community/eslint-plugin-jest/blob/v29.16.1/docs/rules/prefer-todo.md)
- [Source code](https://github.com/jest-community/eslint-plugin-jest/blob/v29.16.1/src/rules/prefer-todo.ts)
