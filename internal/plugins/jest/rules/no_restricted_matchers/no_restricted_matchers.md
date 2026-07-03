# no-restricted-matchers

## Rule Details

Disallow specific Jest matchers and modifiers in `expect()` chains. Use this rule to ban matchers or modifiers that your team prefers to avoid, and optionally suggest alternatives via custom messages.

Restrictions are matched against the **start** of an `expect()` chain. For example, banning `not` reports any chain that begins with `.not`, while banning `not.toBe` only reports that specific prefix. To ban a matcher in every form (with or without `.not`, `.resolves`, or `.rejects`), list each permutation you want to disallow.

By default, no matchers or modifiers are restricted.

Examples of **incorrect** code for this rule with the following configuration:

```json
{
  "jest/no-restricted-matchers": [
    "error",
    {
      "toBeFalsy": null,
      "resolves": "Use `expect(await promise)` instead.",
      "toHaveBeenCalledWith": null,
      "not.toHaveBeenCalledWith": null,
      "resolves.toHaveBeenCalledWith": null,
      "rejects.toHaveBeenCalledWith": null,
      "resolves.not.toHaveBeenCalledWith": null,
      "rejects.not.toHaveBeenCalledWith": null
    }
  ]
}
```

```javascript
it('is false', () => {
  // if this has a modifier (i.e. `not.toBeFalsy`), it would be considered fine
  expect(a).toBeFalsy();
});

it('resolves', async () => {
  // all uses of this modifier are disallowed, regardless of matcher
  await expect(myPromise()).resolves.toBe(true);
});

describe('when an error happens', () => {
  it('does not upload the file', async () => {
    // all uses of this matcher are disallowed
    expect(uploadFileMock).not.toHaveBeenCalledWith('file.name');
  });
});
```

## Options

- First argument (required to enable the rule): object whose keys are restricted matcher chains and whose values are custom messages.
  - Keys are dot-separated chains such as `toBe`, `not.toBe`, `resolves.toBe`, or `resolves.not.toBe`.
  - Values are either a string (shown as the diagnostic message) or `null` (uses the default message: ``Use of `{chain}` is restricted``).

Examples of **incorrect** code with `{ "toBe": "Prefer `toStrictEqual` instead" }`:

```javascript
expect(a).toBe(b);
expect(a)['toBe'](b);
```

Examples of **incorrect** code with `{ "not.toBe": null }`:

```javascript
expect(a).not.toBe(b);
```

Examples of **correct** code with `{ "not.toBe": null }`:

```javascript
expect(a).toBe(b);
expect(a).resolves.not.toBe(b);
```

## Original Documentation

- [jest/no-restricted-matchers](https://github.com/jest-community/eslint-plugin-jest/blob/main/docs/rules/no-restricted-matchers.md)
