# no-unneeded-async-expect-function

## Rule Details

Disallow an async function wrapper around a single awaited call to a locally declared async function passed to Jest `.resolves` or `.rejects` assertions.

The rule recognizes global `expect` and renamed imports from `@jest/globals`. It reports only a concise wrapper without parameters, such as `async () => await operation()`, where the zero-argument call resolves to an earlier local `const` initialized with an async arrow function. A block wrapper can intentionally discard a fulfilled value, and a normal function can depend on its dynamic `this` or `arguments`, so neither is reported. Calls with unknown behavior, arguments, explicit type arguments, mutable bindings or member receivers are also not reported.

No option or autofix is available. Reported wrappers can contain comments whose intended placement cannot be preserved reliably by replacing the wrapper with the referenced function.

## Incorrect

```js
const loadUser = async () => ({ name: 'Ada' });

it('loads a user', async () => {
  await expect(async () => await loadUser()).resolves.toEqual({ name: 'Ada' });
});
```

## Correct

```js
it('loads a user', async () => {
  await expect(loadUser).resolves.toEqual({ name: 'Ada' });
});
```

## Original Documentation

- [eslint-plugin-jest: no-unneeded-async-expect-function](https://github.com/jest-community/eslint-plugin-jest/blob/v29.16.0/docs/rules/no-unneeded-async-expect-function.md)
- [Source code](https://github.com/jest-community/eslint-plugin-jest/blob/v29.16.0/src/rules/no-unneeded-async-expect-function.ts)
