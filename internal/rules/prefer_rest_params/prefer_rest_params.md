# prefer-rest-params

## Rule Details

Requires rest parameters instead of `arguments`.

There are rest parameters in ES2015. We can use that feature for variadic functions instead of the `arguments` variable. `arguments` does not have methods of `Array.prototype`, so it's a bit inconvenient.

Examples of **incorrect** code for this rule:

```javascript
function foo() {
  console.log(arguments);
}

function foo(action) {
  var args = Array.prototype.slice.call(arguments, 1);
  action.apply(null, args);
}

function foo(action) {
  var args = [].slice.call(arguments, 1);
  action.apply(null, args);
}
```

Examples of **correct** code for this rule:

```javascript
function foo(...args) {
  console.log(args);
}

function foo(action, ...args) {
  action.apply(null, args);
}

// This is not a use of `arguments` itself
function foo() {
  arguments.length;
  arguments.callee;
}
```

## Original Documentation

- [ESLint prefer-rest-params](https://eslint.org/docs/latest/rules/prefer-rest-params)
