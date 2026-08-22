# prefer-ending-with-an-expect

## Rule Details

Prefer ending a test body with an assertion. A test whose last statement is not
an `expect` (or another configured assert function) often indicates unfinished
work that may pass silently.

Skipped [`test.todo` / `it.todo`](https://jestjs.io/docs/api#testtodoname)
bodies are ignored.

Examples of **incorrect** code for this rule:

```javascript
it('lets me change the selected option', () => {
  const container = render(MySelect, {
    props: { options: [1, 2, 3], selected: 1 },
  });

  expect(container).toBeDefined();
  expect(container.toHTML()).toContain('<option value="1" selected>');

  container.setProp('selected', 2);
});
```

Examples of **correct** code for this rule:

```javascript
it('lets me change the selected option', () => {
  const container = render(MySelect, {
    props: { options: [1, 2, 3], selected: 1 },
  });

  expect(container).toBeDefined();
  expect(container.toHTML()).toContain('<option value="1" selected>');

  container.setProp('selected', 2);

  expect(container.toHTML()).not.toContain('<option value="1" selected>');
  expect(container.toHTML()).toContain('<option value="2" selected>');
});
```

## Options

```ts
interface Options {
  assertFunctionNames?: string[];
  additionalTestBlockFunctions?: string[];
}
```

### `assertFunctionNames`

Names of functions treated as assertions. Patterns follow
[eslint-plugin-jest](https://github.com/jest-community/eslint-plugin-jest/blob/main/docs/rules/prefer-ending-with-an-expect.md):
`*` matches a dot-separated segment; `**` matches zero or more segments. Default
is `["expect"]`. Special regex characters in names may need escaping (for
example `expect\\$`).

Examples of **incorrect** code with `{ "assertFunctionNames": ["expect"] }`:

```json
{
  "jest/prefer-ending-with-an-expect": [
    "error",
    { "assertFunctionNames": ["expect"] }
  ]
}
```

```javascript
import { expectSaga } from 'redux-saga-test-plan';

test('returns sum', () => {
  expectSaga(addSaga, 1, 1).returns(2).run();
});
```

Examples of **correct** code with
`{ "assertFunctionNames": ["expect", "expectSaga"] }`:

```json
{
  "jest/prefer-ending-with-an-expect": [
    "error",
    { "assertFunctionNames": ["expect", "expectSaga"] }
  ]
}
```

```javascript
import { expectSaga } from 'redux-saga-test-plan';

test('returns sum', () => {
  expectSaga(addSaga, 1, 1).returns(2).run();
});
```

Examples of **correct** code for [SuperTest](https://www.npmjs.com/package/supertest)
with `{ "assertFunctionNames": ["expect", "request.**.expect"] }`:

```json
{
  "jest/prefer-ending-with-an-expect": [
    "error",
    { "assertFunctionNames": ["expect", "request.**.expect"] }
  ]
}
```

```javascript
const request = require('supertest');
const express = require('express');

const app = express();

describe('GET /user', function () {
  it('responds with json', function (done) {
    doSomething();

    request(app).get('/user').expect('Content-Type', /json/).expect(200, done);
  });
});
```

### `additionalTestBlockFunctions`

Extra function names treated like `test` / `it` wrappers so their callbacks are
also required to end with an assertion.

Examples of **correct** code with
`{ "additionalTestBlockFunctions": ["each.test"] }`:

```json
{
  "jest/prefer-ending-with-an-expect": [
    "error",
    { "additionalTestBlockFunctions": ["each.test"] }
  ]
}
```

```javascript
each([
  [2, 3],
  [1, 3],
]).test(
  'the selection can change from %d to %d',
  (firstSelection, secondSelection) => {
    const container = render(MySelect, {
      props: { options: [1, 2, 3], selected: firstSelection },
    });

    expect(container).toBeDefined();
    container.setProp('selected', secondSelection);

    expect(container.toHTML()).toContain(
      `<option value="${secondSelection}" selected>`,
    );
  },
);
```

## Original Documentation

- [jest/prefer-ending-with-an-expect](https://github.com/jest-community/eslint-plugin-jest/blob/main/docs/rules/prefer-ending-with-an-expect.md)
