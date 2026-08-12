# expect-expect

## Rule Details

Enforce that every Rstest test registration contains at least one assertion. A
test with no assertion can pass without verifying anything, hiding the fact that
its intent was never checked.

Examples of incorrect code:

```ts
test("does nothing", () => {
  doSomething();
});

it("empty", () => {});
```

Examples of correct code:

```ts
test("asserts", () => {
  expect(value).toBe(1);
});

it("uses assert", () => {
  assert.equal(value, 1);
});

it("in a promise callback", () =>
  loadUser().then((user) => expect(user).toBeDefined()));

it("named callback", run);
function run() {
  expect(value).toBe(1);
}

const checkValue = () => {
  expect(value).toBe(1);
};
test("variable callback", checkValue);
```

`test.todo` / `it.todo` have no callback and are exempt:

```ts
test.todo("later");
```

The rule recognizes Rstest test registrations from globals, `@rstest/core`
imports and aliases, `require`, namespace access, `import.meta.rstest`,
`test.extend(...)`, `.each` / `.for`, and `@rstest/playwright`. `describe` and
lifecycle hooks are not test registrations and are never required to assert.
Named function declarations and variable functions are resolved independently
of whether they appear before or after the test registration.

## Options

```json
{
  "rstest/expect-expect": [
    "warn",
    {
      "assertFunctionNames": ["expect", "assert"],
      "additionalTestBlockFunctions": []
    }
  ]
}
```

### `assertFunctionNames`

Names of functions treated as assertions. Defaults to `["expect", "assert"]`,
matching the two assertion entry points Rstest exposes as globals. Names support
wildcards: `request.*.expect`, `request.**.expect`, `expect*`.

### `additionalTestBlockFunctions`

Additional function names to treat as test blocks.

The rule matches assertions by callee name only; unlike `no-conditional-expect`
it does not resolve the module origin of `expect`. It provides no automatic fix.
