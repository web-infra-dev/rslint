# no-lone-blocks

## Rule Details

This rule disallows nested (non-function) lone blocks. A lone block is a block that is not part of an `if`, `for`, `while`, `function`, `try`, class static block, or other statement that naturally introduces one. In ES6+, a block can be useful to scope `let`, `const`, `class`, `function`, or `using` declarations — this rule only flags blocks that do not contain such block-scoped bindings.

Examples of **incorrect** code for this rule:

```javascript
{}

{
  var x = 1;
}

if (foo) {
  bar();
  {
    baz();
  }
}

function foo() {
  {
    var x = 1;
  }
}

class C {
  static {
    {
      foo();
    }
  }
}
```

Examples of **correct** code for this rule:

```javascript
while (foo) {
  bar();
}

if (foo) {
  if (bar) {
    baz();
  }
}

{
  let x = 1;
}

{
  const y = 2;
}

{
  class Bar {}
}

switch (foo) {
  case bar: {
    baz();
  }
}

class C {
  static {
    foo();
  }
}
```

## Original Documentation

- [ESLint rule: no-lone-blocks](https://eslint.org/docs/latest/rules/no-lone-blocks)
