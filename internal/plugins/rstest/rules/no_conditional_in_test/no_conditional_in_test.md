# no-conditional-in-test

## Rule Details

Disallow conditional logic in Rstest test bodies. A conditional usually means
that one test is covering multiple execution paths, which makes it harder to
see which behavior the test is intended to verify. Prefer a separate test for
each branch.

Examples of **incorrect** code for this rule:

```ts
test('loads the user', () => {
  if (enabled) {
    loadUser();
  }
});

it('renders a mode', () => {
  switch (mode) {
    case 'none':
      renderNone();
      break;
    case 'full':
      renderFull();
      break;
  }
});
```

Examples of **correct** code for this rule:

```ts
describe('user flow', () => {
  if (enabled) {
    test('loads the user', () => {
      loadUser();
    });
  }
});

beforeEach(() => {
  switch (mode) {
    case 'none':
      renderNone();
      break;
    case 'full':
      renderFull();
      break;
  }
});

function pickLabel(kind: string) {
  return kind === 'full' ? 'Full' : 'Compact';
}

test('renders a label', () => {
  expect(pickLabel(mode)).toBe('Full');
});
```

Conditionals inside `describe` blocks, hooks, and helper functions declared
outside a test are not reported. Conditionals in helper functions declared
inside a test are reported because they are still part of that test body.

## Options

- First argument (optional): object with `allowOptionalChaining`
  - `allowOptionalChaining`: whether optional chaining (`?.`) is allowed inside
    test bodies. Default is `true`.

When `allowOptionalChaining` is `false`, optional property access, element
access, and calls are also reported:

```json
{
  "rstest/no-conditional-in-test": [
    "error",
    {
      "allowOptionalChaining": false
    }
  ]
}
```

Examples of **incorrect** code with `{ "allowOptionalChaining": false }`:

```ts
test('loads a value', () => {
  const value = api?.result;
});

test('calls a method', () => {
  client?.run();
});
```

Examples of **correct** code with `{ "allowOptionalChaining": false }`:

```ts
test('loads a value', () => {
  const value = api!.result;
});
```

## Limitations

The scope is the test registration call itself, reached through supported
Rstest forms: global `test` / `it`, imports from `@rstest/core` and
`@rstest/playwright`, namespace and CommonJS access, `import.meta.rstest`, and
parameterized `.each` / `.for`. Everything written inside that call is checked,
including the title, the options and timeout arguments, the `.each` data, and a
callback passed through a wrapper such as `test('case', wrap(() => {}))`.

A function declared outside the call is not checked, even when the call names
it as its callback: in `test('case', callback); function callback() {}` the
body of `callback` is outside the test call, so conditionals in it are not
reported.

Unlike the upstream rule, an inner registration exiting does not clear the
outer test's scope, so a conditional written after a nested `test(...)` is
still reported.

## Original Documentation

- [eslint-plugin-jest: no-conditional-in-test](https://github.com/jest-community/eslint-plugin-jest/blob/v29.16.1/docs/rules/no-conditional-in-test.md)
- [Source code](https://github.com/jest-community/eslint-plugin-jest/blob/v29.16.1/src/rules/no-conditional-in-test.ts)
