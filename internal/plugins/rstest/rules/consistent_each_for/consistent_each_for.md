# consistent-each-for

## Rule Details

Rstest offers two ways to register a parameterized test or suite, and they are not interchangeable: [`.each`](https://rstest.rs/api/runtime-api/test-api/test#testeach) spreads an array case into separate callback parameters, while [`.for`](https://rstest.rs/api/runtime-api/test-api/test#testfor) passes the case through untouched and hands the callback the [TestContext](https://rstest.rs/api/runtime-api/test-api/test#testcontext) as a second argument. A file that uses both makes every parameterized block a small puzzle about which shape its callback receives. This rule picks one form per registration name and reports the other.

`test`, `it`, and `describe` are configured independently, and a name that is not configured is not checked. That means the rule does nothing at all until at least one preference is set — enabling it without options is not an error, it simply never reports.

Both call forms are covered: a case array, as in `test.each([1, 2])`, and a tagged template table, as in `` test.each`a | b` ``. Modifiers written between the registration and the accessor — `.skip`, `.only`, `.concurrent`, `.skipIf(…)`, `.extend({…})` — do not hide it. The registration is resolved rather than name-matched, so an import from `@rstest/core` or `@rstest/playwright`, a global, and a local alias are all recognised, while a same-named import from another test framework and a local variable that merely shares the name are left alone. A suite registered as `test.describe` is governed by the `describe` preference. The diagnostic points at the `.each` or `.for` accessor, or, when a variable holds the already-parameterized registration, at that variable.

No fix is offered. Whether switching forms requires rewriting the callback depends on the shape of the cases, which is a runtime value the rule cannot inspect: `[[1, 2]]` reaches an `.each` callback as `(a, b)` and a `.for` callback as `([a, b])`, while a list of scalars or a template table reaches both the same way.

## Incorrect

```ts
// { "test": "for" }
test.each([
  [2, 4],
  [3, 9],
])('squares %i', (input, expected) => {
  expect(square(input)).toBe(expected);
});
```

## Correct

```ts
// { "test": "for" }
test.for([
  [2, 4],
  [3, 9],
])('squares %i', ([input, expected]) => {
  expect(square(input)).toBe(expected);
});
```

## Options

```json
{
  "rstest/consistent-each-for": [
    "error",
    {
      "test": "for",
      "describe": "each"
    }
  ]
}
```

| Option | Type | Default | Description |
| --- | --- | --- | --- |
| `describe` | `string` | none | Parameterized form to require for `describe`, either `"each"` or `"for"`. |
| `it` | `string` | none | Parameterized form to require for `it`, either `"each"` or `"for"`. |
| `test` | `string` | none | Parameterized form to require for `test`, either `"each"` or `"for"`. |
