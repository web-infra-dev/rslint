# no-const-assign

## Rule Details

Disallows reassigning variables declared with `const`, `using`, or `await using`. These constant bindings cannot be reassigned after their initial declaration. Attempting to modify them (via assignment, increment, decrement, or destructuring assignment) is always a mistake and would throw a `TypeError` at runtime.

Examples of **incorrect** code for this rule:

```javascript
const x = 1;
x = 2;

const y = 0;
y++;

const { z } = obj;
z = 3;

const [a, b] = [1, 2];
[a, b] = [3, 4];

using resource = acquireResource();
resource = acquireOtherResource();
```

Examples of **correct** code for this rule:

```javascript
const x = 1;
console.log(x);

let y = 0;
y = 1;

const obj = {};
obj.key = 'value'; // mutating properties is fine

using resource = acquireResource();
resource.use();
```

## Original Documentation

- [ESLint no-const-assign](https://eslint.org/docs/latest/rules/no-const-assign)
