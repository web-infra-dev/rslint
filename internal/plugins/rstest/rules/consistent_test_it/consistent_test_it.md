# consistent-test-it

## Rule Details

Enforce a consistent choice of `test` and `it` when registering tests. By default, use `test` outside `describe` calls and `it` inside them. Setting `fn` also sets the preference inside `describe`, unless `withinDescribe` overrides it.

The rule recognizes Rstest globals, imports from `@rstest/core` and `rstack/test`, namespace imports, destructured `require` imports, and `import.meta.rstest`. Import aliases and same-file `const` aliases named `test` or `it` are checked by that local name; other aliases use their original API name. Nested and parameterized suites count as `describe` scopes; this is based on where a call is written, so a separately declared callback does not acquire the scope of the suite that uses it. The rule does not require type information.

Test modifiers, conditional registration, `.extend`, and array or tagged-template `.each` and `.for` registrations are checked. See the [Rstest test API](https://rstest.rs/api/runtime-api/test-api/test). Local functions and imports from other modules are ignored, as are dynamic property names and calls wrapped in TypeScript assertions or non-null assertions. `@rstest/playwright` tests are exempt because that module has no `it` export. Imports and references that do not register a test are not violations.

## Incorrect

```ts
import { describe, it, test, expect } from '@rstest/core';

it('loads the default configuration', () => {
  expect(loadConfig()).toEqual(defaultConfig);
});

describe('configuration overrides', () => {
  test('uses the requested port', () => {
    expect(loadConfig({ port: 8080 }).port).toBe(8080);
  });
});
```

## Correct

```ts
import { describe, it, test, expect } from '@rstest/core';

test('loads the default configuration', () => {
  expect(loadConfig()).toEqual(defaultConfig);
});

describe('configuration overrides', () => {
  it('uses the requested port', () => {
    expect(loadConfig({ port: 8080 }).port).toBe(8080);
  });
});
```

## Options

```json
{
  "rstest/consistent-test-it": [
    "error",
    {
      "fn": "it",
      "withinDescribe": "test"
    }
  ]
}
```

| Option | Type | Default | Description |
| ------ | ---- | ------- | ----------- |
| `fn` | `string` | `"test"` | Use `"test"` or `"it"` outside suites and inside suites without an explicit override. |
| `withinDescribe` | `string` | `"it"` | Use `"test"` or `"it"` inside suites; when omitted, use the explicitly configured `fn`, or `"it"` if neither option is set. |

## Autofix

The fix replaces the test API name and preserves its modifiers, arguments, comments, and parameter tables. Named imports gain the preferred API if needed; the original import remains available for other references. Namespace imports and direct `import.meta.rstest` access keep their receiver and property quoting.

Calls through renamed imports, local aliases, or CommonJS bindings without an existing preferred named import are reported without a fix. Local namespace aliases and CommonJS namespace calls are also reported without a fix. A fix that introduces an identifier is withheld when that name would collide with or be captured by another binding.
