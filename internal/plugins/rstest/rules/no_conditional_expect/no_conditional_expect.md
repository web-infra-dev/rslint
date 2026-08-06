# no-conditional-expect

## Rule Details

Disallow calling Rstest `expect` APIs conditionally. When an assertion only
runs on one branch, the test can pass without executing that assertion.

Examples of incorrect code:

```ts
test("loads user", () => {
  if (user) {
    expect(user.name).toBe("Alice");
  }
});

test("rejects", async () => {
  await request().catch((error) => {
    expect(error).toBeInstanceOf(Error);
  });
});

test("uses local expect", ({ expect }) => {
  ready && expect(value).toBeDefined();
});
```

The rule recognizes Rstest APIs from globals, `@rstest/core`,
`import.meta.rstest`, test context `expect`, Browser Mode `expect.element`,
and `@rstest/playwright`.

Examples of correct code:

```ts
test("loads user", () => {
  expect(user).toBeDefined();
  expect(user?.name).toBe("Alice");
});

test("rejects", async () => {
  await expect(request()).rejects.toThrow();
});

test("prepares conditionally", () => {
  if (mode === "a") {
    setupA();
  } else {
    setupB();
  }

  expect(result).toBeDefined();
});
```

The rule has no options and does not provide an automatic fix.
