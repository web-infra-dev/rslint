# return-await

## Rule Details

Enforce consistent returning of awaited values. This rule controls whether `return await` should be used in async functions. Depending on the configuration, it enforces one of several strategies:

- `in-try-catch` (default): Requires `await` in try/catch/finally blocks (for proper error handling) and disallows it elsewhere (for performance).
- `always`: Always requires `return await`.
- `never`: Always disallows `return await`.
- `error-handling-correctness-only`: Only requires `await` where it affects error handling correctness, with no preference otherwise.

Using `return await` inside try/catch ensures the promise rejection is caught in the local catch block. Outside try/catch, the `await` adds unnecessary overhead.

Examples of **incorrect** code for this rule (with default `in-try-catch`):

```typescript
async function foo() {
  return await bar(); // unnecessary await outside try/catch
}

async function baz() {
  try {
    return promise; // missing await inside try
  } catch (e) {
    handleError(e);
  }
}
```

Examples of **correct** code for this rule (with default `in-try-catch`):

```typescript
async function foo() {
  return bar();
}

async function baz() {
  try {
    return await promise;
  } catch (e) {
    handleError(e);
  }
}
```

## Original Documentation

- [typescript-eslint return-await](https://typescript-eslint.io/rules/return-await)
