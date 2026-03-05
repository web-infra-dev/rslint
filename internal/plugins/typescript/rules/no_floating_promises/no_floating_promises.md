# no-floating-promises

## Rule Details

Requires Promise-like statements to be handled appropriately. A "floating" promise is one that is created without any code to handle potential errors. Floating promises can cause unexpected behavior because errors in them will be silently ignored. This rule reports on promises in expression statements that are not awaited, not chained with `.catch()` or `.then()` with a rejection handler, and not explicitly ignored with the `void` operator.

Examples of **incorrect** code for this rule:

```typescript
async function fetchData() {
  fetch('https://example.com');
}

const promise = new Promise(resolve => resolve('value'));
promise;

async function run() {
  doAsyncWork();
}
```

Examples of **correct** code for this rule:

```typescript
async function fetchData() {
  await fetch('https://example.com');
}

const promise = new Promise(resolve => resolve('value'));
await promise;

promise.catch(handleError);

void promise;
```

## Original Documentation

- [typescript-eslint no-floating-promises](https://typescript-eslint.io/rules/no-floating-promises)
