# require-test-timeout

## Rule Details

Requires every test to declare a timeout instead of relying on the project-wide default. An explicit budget makes a slow test a visible decision and stops a hang from running for the whole default.

A timeout can come from the test itself, from an enclosing suite, or from the runtime configuration. `test('name', fn, 5000)` and `test('name', { timeout: 5000 }, fn)` both declare one, and `0` counts because it disables the timeout, while a negative value does not. A [suite](https://rstest.rs/api/runtime-api/test-api/describe) timeout applies to the tests nested inside it, and `rs.setConfig({ testTimeout })` covers the tests after it until `resetConfig()` restores the [default](https://rstest.rs/config/test/test-timeout).

Globals, imports, aliases, namespace members, `import.meta.rstest`, parameterized tests, named callbacks, and Playwright integrations are recognized. `test.todo`, `test.skip`, and registrations without a callback are exempt. A timeout the rule cannot resolve statically, such as an imported constant or a spread into the options object, counts as declared, as does a test inside a function that cannot be tied to a suite.

## Incorrect

```ts
test('imports the catalog', async () => {
  await importCatalog(fixture);
});

describe('catalog export', () => {
  it('writes every row', async () => {
    await exportCatalog(destination);
  });
});
```

## Correct

```ts
test('imports the catalog', async () => {
  await importCatalog(fixture);
}, 30_000);

describe('catalog export', { timeout: 30_000 }, () => {
  it('writes every row', async () => {
    await exportCatalog(destination);
  });
});
```
