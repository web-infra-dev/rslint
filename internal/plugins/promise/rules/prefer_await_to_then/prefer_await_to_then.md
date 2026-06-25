# prefer-await-to-then

## Rule Details

Prefer `async`/`await` over `.then()`/`.catch()`/`.finally()` for consuming Promise values.

Examples of **incorrect** code for this rule:

```javascript
function foo() {
  fetchData().then((data) => {
    process(data);
  });
}

function bar() {
  fetchData()
    .then((data) => data.json())
    .catch((err) => console.error(err));
}
```

Examples of **correct** code for this rule:

```javascript
async function foo() {
  const data = await fetchData();
  process(data);
}

async function bar() {
  try {
    const data = await fetchData();
    return await data.json();
  } catch (err) {
    console.error(err);
  }
}
```

By default, `.then()`/`.catch()`/`.finally()` inside an `await` or `yield` expression, or inside a class constructor, are not reported — those positions cannot trivially be rewritten with `await`.

Examples of **correct** code for this rule in default (non-strict) mode:

```javascript
async function foo() {
  // await wraps the .then() — not reported
  return await thing().then();
}

function* gen() {
  // yield wraps the .then() — not reported
  yield thing().then();
}

class Foo {
  constructor() {
    // inside a constructor — not reported
    doSomething.then(cb);
  }
}
```

### `strict` option

When `strict: true`, the rule also reports `.then()`/`.catch()`/`.finally()` inside `await`/`yield` expressions and constructors.

```json
{ "promise/prefer-await-to-then": ["error", { "strict": true }] }
```

```javascript
// reported with strict: true
async function foo() {
  return await thing().then();
}
```

## Original Documentation

https://github.com/eslint-community/eslint-plugin-promise/blob/main/docs/rules/prefer-await-to-then.md
