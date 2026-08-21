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

The rule follows Rstest globals, named imports, ES module namespace imports,
`import.meta.rstest` and same-file `const` aliases only when the alias chain is
provably a bare test API reference. It intentionally fails closed for aliases
that hide modifiers or factories such as `.skip`, `.fails`, `.each(...)`,
`.for(...)` or `.extend(...)`.

Whole-module CommonJS namespace objects, such as
`const core = require('@rstest/core')`, are left unchanged. Their API properties
are mutable even when the namespace binding is declared with `const`, so the
rule cannot safely assume that `core.test` still refers to Rstest. Repeated
`.skip` chains are also left unchanged because replacing only one accessor
would leave the test skipped.

The options overload is recognized only when the second argument is written as
an object literal. Calls such as `test('case', options)` and
`test('case', options, () => {})` are left unchanged because the identifier may
still be the callback.

Computed identifier access such as `test[skip]('case', () => {})` is left
unchanged. The parser does not treat the computed identifier as a statically
known `.skip` member, so the rule intentionally fails closed: it neither
reports nor autofixes that form.

## Original Documentation

- [eslint-plugin-jest: prefer-todo](https://github.com/jest-community/eslint-plugin-jest/blob/v29.16.1/docs/rules/prefer-todo.md)
- [Source code](https://github.com/jest-community/eslint-plugin-jest/blob/v29.16.1/src/rules/prefer-todo.ts)
