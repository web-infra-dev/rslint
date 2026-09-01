# consistent-test-filename

## Rule Details

Requires every test file to be named according to a single convention. Rstest runs both `*.test.*` and `*.spec.*` files by default, so a project that mixes the two makes tests harder to locate and forces every coverage glob, CI shard, and editor filter to spell out both forms.

`allTestPattern` decides which files count as test files, and `pattern` is the convention those files have to follow. Under the defaults a `*.spec.*` file is a test file whose name does not match the required `*.test.*` form, so it is reported. Both defaults accept the `.cts`, `.mts`, `.cjs`, and `.mjs` extensions that Rstest's [`include`](https://rstest.rs/config/test/include) glob accepts by default.

Only the path is examined, never the contents. A file named `*.spec.ts` is reported even when it registers no test, and a source file carrying [in-source tests](https://rstest.rs/config/test/include-source) is never reported, because its name is an ordinary source file's name. Both patterns are matched against the whole path rather than the basename, so a pattern such as `__tests__` can select an entire directory. The diagnostic is reported at the start of the file.

## Incorrect

```ts
// user-service.spec.ts
test('creates a user', () => {
  expect(createUser({ name: 'Ada' })).toMatchObject({ name: 'Ada' });
});
```

## Correct

```ts
// user-service.test.ts
test('creates a user', () => {
  expect(createUser({ name: 'Ada' })).toMatchObject({ name: 'Ada' });
});
```

## Options

```json
{
  "rstest/consistent-test-filename": [
    "error",
    {
      "pattern": ".*\\.spec\\.(c|m)?[tj]sx?$"
    }
  ]
}
```

| Option | Type | Default | Description |
| --- | --- | --- | --- |
| `pattern` | `string` | `.*\.test\.(c\|m)?[tj]sx?$` | Regular expression the path of a test file must match. |
| `allTestPattern` | `string` | `.*\.(test\|spec)\.(c\|m)?[tj]sx?$` | Regular expression selecting the paths the rule checks. |
