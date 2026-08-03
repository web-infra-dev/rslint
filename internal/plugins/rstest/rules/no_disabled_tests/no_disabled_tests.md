# no-disabled-tests

## Rule Details

Disallow explicitly skipped or incomplete Rstest tests. This rule reports
`test.skip`, `it.skip`, and `describe.skip`, including valid modifier,
parameterized, and `test.extend()` chains. It also reports `test` and `it`
registrations that omit the callback.

Examples of **incorrect** code for this rule:

```ts
describe.skip('suite', () => {});
it.skip('case', () => {});
test.skip.each(rows)('case $value', (value) => {});
test('missing callback');
test('options without a callback', { timeout: 100 });
```

Examples of **correct** code for this rule:

```ts
describe('suite', () => {});
test('case', () => {});
test.todo('future case');

test('runtime condition', (context) => {
  if (!available()) {
    context.skip();
  }
});
```

`test.todo` is an explicit, visible todo state and is not reported.
`context.skip()` is Rstest's runtime control-flow API and is also not reported.
Conditional `skipIf` and `runIf` expressions are not evaluated by this rule.

## Limitations

The rule follows Rstest APIs imported from `@rstest/core`, globals, namespace
imports, CommonJS requires, and same-file `const` aliases. It does not trace
custom test APIs through arbitrary cross-file re-exports.

The missing-callback check covers Rstest's `(description, options, fn?)`
overload only when the options argument is written as an object literal. An
indirect options call such as `test('case', options)` is not reported, because
the identifier may resolve to the callback itself.

This rule does not provide an autofix because removing `skip` or inventing a
callback would change test behavior.
