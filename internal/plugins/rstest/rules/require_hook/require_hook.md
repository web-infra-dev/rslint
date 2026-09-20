# require-hook

## Rule Details

Requires setup and teardown code to be written inside a lifecycle hook rather than directly in a test file's module scope or in a `describe` callback.

Rstest collects a test file by executing it. Module-level statements and every `describe` callback run during that collection pass, strictly before any `beforeAll` or `beforeEach` runs, and they run even for a suite that is skipped — `describe.skip` still executes its callback while the file is collected. Setup written there therefore happens whether or not the tests that need it run, happens in file order rather than in the order the suites declare, and has no matching teardown point, since `afterAll` cannot undo work done before the run began.

Moving the work into `beforeAll`, `beforeEach`, `afterEach`, or `afterAll` ties it to the tests that need it and gives it a place to be reverted.

The rule reports two kinds of statement at the top level of a file or of a `describe` callback:

- a function call that is not part of Rstest's own API;
- a `let` or `var` declaration with an initializer other than `null` or `undefined`.

A `const` declaration is left alone, as are Rstest's own calls: test and suite registrations, lifecycle hooks, `expect` chains, and calls on the `rs` / `rstest` utilities object, including through a renamed import, a namespace import, or `import.meta.rstest`. A `describe` callback is only inspected when it is written at the call site; a callback passed by name is not, because the same function may be registered more than once.

## Incorrect

```ts
import { describe, it, expect } from '@rstest/core';

initializeCityDatabase();

describe('cities', () => {
  let consoleWarnSpy = rs.spyOn(console, 'warn');

  it('has Vienna', () => {
    expect(isCity('Vienna')).toBe(true);
  });
});
```

## Correct

```ts
import { beforeEach, describe, it, expect } from '@rstest/core';

beforeEach(() => {
  initializeCityDatabase();
});

describe('cities', () => {
  let consoleWarnSpy;

  beforeEach(() => {
    consoleWarnSpy = rs.spyOn(console, 'warn');
  });

  it('has Vienna', () => {
    expect(isCity('Vienna')).toBe(true);
  });
});
```

## Options

```json
{
  "rstest/require-hook": [
    "error",
    {
      "allowedFunctionCalls": ["enableAutoDestroy"]
    }
  ]
}
```

| Option | Type | Default | Description |
| --- | --- | --- | --- |
| `allowedFunctionCalls` | `string[]` | `[]` | Names of function calls that may stand outside a hook. A name is matched against the whole callee chain, so `helper.setup` allows `helper.setup()` but not `setup()`. |
