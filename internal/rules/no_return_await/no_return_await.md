# no-return-await

## Rule Details

Disallows `return await` inside an `async` function. Awaiting a promise only to hand it straight back to the caller adds an extra microtask tick without changing the resolved value, so the `await` can be dropped.

The `await` is only redundant when it sits on the branch that actually produces the returned value — the argument of a `return`, the body of a concise arrow function, either branch of a conditional, the right side of `&&` / `||` / `??`, or the last expression of a comma sequence.

An `await` inside a `try` block, or inside a `catch` clause that is followed by `finally`, is **not** reported: removing it there would move rejection handling outside the enclosing error handler and change control flow.

Each report carries a suggestion that removes the `await` keyword. The suggestion is omitted when the keyword and its operand sit on different lines, since removing it would join the two lines.

Examples of **incorrect** code for this rule:

```javascript
async function foo() {
  return await bar();
}

async function baz() {
  return (a && (await bar()));
}

async function qux() {
  return cond ? await bar() : b;
}

const quux = async () => await bar();
```

Examples of **correct** code for this rule:

```javascript
async function foo() {
  const result = await bar();
  return result;
}

async function baz() {
  return bar();
}

async function qux() {
  return (await bar()) || a;
}

async function quux() {
  try {
    return await bar();
  } catch (e) {
    handle(e);
  }
}
```

## Original Documentation

- [ESLint no-return-await](https://eslint.org/docs/latest/rules/no-return-await)
