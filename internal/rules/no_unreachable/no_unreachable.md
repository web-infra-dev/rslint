# no-unreachable

## Rule Details

Disallows unreachable code after `return`, `throw`, `break`, and `continue` statements. Because these statements unconditionally exit a block of code, any statements after them cannot be executed and are therefore unreachable.

Function declarations are allowed after terminal statements because they are hoisted. Similarly, `var` declarations without initializers are allowed because the declaration itself is hoisted, even though the assignment would be unreachable.

Examples of **incorrect** code for this rule:

```javascript
function foo() {
  return true;
  console.log('done');
}

function bar() {
  throw new Error('oops');
  console.log('done');
}

while (value) {
  break;
  console.log('done');
}

while (value) {
  continue;
  console.log('done');
}

function baz() {
  return;
  var x = 1;
}
```

Examples of **correct** code for this rule:

```javascript
function foo() {
  return bar();
  function bar() {
    return 1;
  }
}

function baz() {
  return;
  var x;
}

function qux() {
  if (condition) {
    return;
  }
  doSomething();
}
```

## Original Documentation

- [ESLint no-unreachable](https://eslint.org/docs/latest/rules/no-unreachable)
