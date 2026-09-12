# no-unnecessary-assertion

## Rule Details

Disallows assertions that a subject is null, undefined, or NaN when its TypeScript type makes that result impossible. These assertions either add no useful coverage or indicate that the declared type does not describe the runtime behavior accurately. The rule requires type information and reports a configuration warning when `strictNullChecks` is disabled.

The rule checks `toBeNull()`, `toBeUndefined()`, `toBeDefined()`, and `toBeNaN()`, together with Chai's `.null`, `.undefined`, and `.NaN` property assertions. In a Chai chain containing multiple assertions, only the first matcher is compared with the original subject type because an earlier matcher can change the value used by later assertions. The rule recognizes ordinary and soft Rstest assertions from globals, named and renamed imports, namespace and `require` bindings, `import.meta.rstest`, and the `expect` supplied by a [test context](https://rstest.rs/api/runtime-api/test-api/test#testcontext).

Assertions using `resolves` or `rejects` are excluded because they inspect a promise result rather than the subject's direct type. For [`expect.poll()`](https://rstest.rs/api/runtime-api/test-api/expect#expectpoll) method assertions, the rule checks the callback's awaited return type and skips the assertion if that type cannot be determined; Chai property assertions after `expect.poll()` are excluded because Rstest evaluates those getters before polling the callback. Browser-only [`expect.element()`](https://rstest.rs/api/runtime-api/test-api/expect#expectelement-browser-mode) assertions are also excluded. An assertion on `any`, `unknown`, `void`, an unresolved generic type, or a union containing the tested type is retained. `void` remains unchecked because TypeScript permits a value-returning function to be assigned to a `void`-returning signature.

## Incorrect

```ts
declare function loadProfile(): { username: string; attempts: string };

test('loads profile metadata', () => {
  const profile = loadProfile();

  expect(profile.username).toBeDefined();
  expect(profile.attempts).not.toBeNaN();
  expect.soft(profile.username).to.not.be.null;
});
```

## Correct

```ts
declare function loadProfile(): {
  username: string | null | undefined;
  attempts: number;
};
declare function fetchUsername(): Promise<string | undefined>;
declare function readUsername(): Promise<string | null>;

test('loads profile metadata', async () => {
  const profile = loadProfile();

  expect(profile.username).toBeDefined();
  expect(profile.attempts).not.toBeNaN();
  expect.soft(profile.username).to.not.be.null;
  await expect(fetchUsername()).resolves.toBeDefined();
  await expect.poll(readUsername).toBeNull();
});
```
