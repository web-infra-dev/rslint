# no-misused-promises

## Rule Details

Disallows Promises in places that are not designed to handle them. This rule catches common mistakes such as using a Promise in a conditional check (where it always evaluates to truthy), passing an async function as a callback where a void return is expected, spreading a Promise in an object, or returning a Promise-returning function where a void return is expected. The rule checks conditionals, void returns (arguments, properties, variables, inherited methods, attributes, and return statements), and object spreads.

Examples of **incorrect** code for this rule:

```typescript
// Promise used in conditional (always truthy)
if (fetchData()) {
}

// Async function passed where void callback expected
[1, 2, 3].forEach(async n => {
  await doSomething(n);
});

// Spreading a Promise in an object
const obj = { ...fetchData() };

// Returning async function where void expected
const listeners = {
  onClick: async () => await handleClick(),
};
```

Examples of **correct** code for this rule:

```typescript
if (await fetchData()) {
}

for (const n of [1, 2, 3]) {
  await doSomething(n);
}

const data = await fetchData();
const obj = { ...data };

const listeners = {
  onClick: () => {
    void handleClick();
  },
};
```

## Original Documentation

- [typescript-eslint no-misused-promises](https://typescript-eslint.io/rules/no-misused-promises)
