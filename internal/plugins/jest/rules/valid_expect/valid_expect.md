# valid-expect

## Rule Details

Enforce valid `expect()` usage in Jest tests. This rule ensures that every
`expect()` call has the expected number of arguments, ends with a matcher that
is actually invoked, uses only supported modifiers (`.not`, `.resolves`,
`.rejects`), and that async assertions are properly `await`ed or returned.

The rule is fixable: when an async assertion appears inside a function, rslint
can insert `async` on the enclosing function and `await` on the assertion.

Examples of **incorrect** code for this rule:

```javascript
// Missing matcher or matcher not called
expect();
expect("something");
expect(true).toBeDefined;
expect(true).resolves;

// Invalid modifier chain
expect(true).nope.toBeDefined();
expect(true).not.resolves.toBeDefined();

// Wrong number of arguments (default: exactly one)
expect();
expect("something", "else");

// Async assertions not awaited or returned
expect(Promise.resolve("Hi!")).resolves.toBe("Hi!");

test("example", () => {
  expect(Promise.resolve(2)).resolves.toBeDefined();
});

test("example", () => {
  Promise.all([
    expect(Promise.resolve("hello")).resolves.toEqual("hello"),
    expect(Promise.resolve("hi")).resolves.toEqual("hi"),
  ]);
});
```

Examples of **correct** code for this rule:

```javascript
expect("something").toEqual("something");
expect([1, 2, 3]).toEqual([1, 2, 3]);
expect(true).toBeDefined();
expect(undefined).not.toBeDefined();

test("example", async () => {
  await expect(Promise.resolve("hello")).resolves.toEqual("hello");
});

test("example", () => {
  return expect(Promise.resolve(2)).resolves.toBeDefined();
});

test("example", async () => {
  await Promise.all([
    expect(Promise.resolve("hello")).resolves.toEqual("hello"),
    expect(Promise.resolve("hi")).resolves.toEqual("hi"),
  ]);
});

// Arrow functions with implicit return are allowed when alwaysAwait is false
test("example", () => expect(Promise.resolve(2)).resolves.toBeDefined());
```

## Options

- `alwaysAwait` (default: `false`): When `true`, only `await` is accepted for
  async assertions inside block statements. `return expect(...).resolves...` is
  reported. One-line arrow functions with an implicit return remain valid.

- `asyncMatchers` (default: `["toResolve", "toReject"]`): Matcher names that
  should be treated as async even without `.resolves` or `.rejects`. Use this
  for custom matchers from libraries such as `jest-extended`.

- `minArgs` (default: `1`): Minimum number of arguments allowed in `expect()`.
  Set to `0` when using libraries such as
  [`jest-expect-message`](https://www.npmjs.com/package/jest-expect-message).

- `maxArgs` (default: `1`): Maximum number of arguments allowed in `expect()`.

Examples of **incorrect** code for `{ "alwaysAwait": true }`:

```javascript
test("example", async () => {
  await expect(Promise.resolve(2)).resolves.toBeDefined();
  return expect(Promise.resolve(1)).resolves.toBe(1);
});
```

Examples of **correct** code for `{ "alwaysAwait": true }`:

```javascript
test("example", async () => {
  await expect(Promise.resolve(2)).resolves.toBeDefined();
  await expect(Promise.resolve(1)).resolves.toBe(1);
});

test("example", () => expect(Promise.resolve(2)).resolves.toBeDefined());
```

Examples of **correct** code for `{ "minArgs": 0 }`:

```javascript
expect().pass();
```

## Original Documentation

- [jest/valid-expect](https://github.com/jest-community/eslint-plugin-jest/blob/main/docs/rules/valid-expect.md)
