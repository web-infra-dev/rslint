# no-await-in-loop

## Rule Details

Disallows `await` expressions inside loop bodies (`for`, `for-in`, `for-of`, `while`, `do-while`). Using `await` in a loop is usually a sign that the program is not taking full advantage of the parallelization benefits of `async`/`await`, as each iteration waits for the previous one to complete. The operations can often be refactored to use `Promise.all()` instead.

This rule allows `await` in `for-await-of` loops, since those are designed to work with async iterables. However, a `for-await-of` nested inside another loop will be flagged.

Examples of **incorrect** code for this rule:

```javascript
async function foo(things) {
  for (const thing of things) {
    await bar(thing);
  }
}

async function foo(things) {
  while (things.length) {
    await bar(things.pop());
  }
}
```

Examples of **correct** code for this rule:

```javascript
async function foo(things) {
  await Promise.all(things.map(thing => bar(thing)));
}

async function foo(things) {
  for await (const thing of asyncIterable) {
    console.log(thing);
  }
}
```

## Original Documentation

- [ESLint no-await-in-loop](https://eslint.org/docs/latest/rules/no-await-in-loop)
