# no-identical-title

## Rule Details

Disallow the same static title for two tests or two `describe` blocks in the
same scope.

Examples of incorrect code:

```ts
test("loads user", () => {});
test.only("loads user", () => {});

describe("api", () => {});
describe.skip("api", () => {});
```

Examples of correct code:

```ts
describe("api", () => {
  test("loads user", () => {});

  describe("nested", () => {
    test("loads user", () => {});
  });
});
```

Test titles and describe titles are tracked separately. Dynamic titles are
ignored. Parameterized `.each` and `.for` registration titles are also ignored
because Rstest expands them at runtime.
