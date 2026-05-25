# no-func-assign

## Rule Details

Disallows reassigning variables that were declared as function declarations. Reassigning a function declaration is almost always a mistake, as it overwrites the function with a different value. This rule checks for assignments, increment/decrement operations, and destructuring assignments that target a function name.

Examples of **incorrect** code for this rule:

```javascript
function foo() {}
foo = bar;

function foo() {}
foo += 1;

function foo() {}
[foo] = arr;
```

Examples of **correct** code for this rule:

```javascript
function foo() {}
foo();

var foo = function () {};
foo = bar; // foo is a variable, not a function declaration

function foo(foo) {
  foo = bar; // foo is a parameter, not the function
}

function foo() {
  var foo = bar; // foo is a local variable, not the function
}
```

## Original Documentation

- [ESLint no-func-assign](https://eslint.org/docs/latest/rules/no-func-assign)
