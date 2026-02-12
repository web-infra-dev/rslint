# no-case-declarations

## Rule Details

Disallow lexical declarations in case clauses. Lexical declarations (`let`, `const`, `function`, `class`) in case clauses are visible in the entire switch block but only get initialized when assigned, which can lead to unexpected behavior.

Examples of **incorrect** code for this rule:

```javascript
switch (foo) {
  case 1:
    let x = 1;
    break;
  case 2:
    const y = 2;
    break;
  case 3:
    function f() {}
    break;
  case 4:
    class C {}
    break;
}
```

Examples of **correct** code for this rule:

```javascript
switch (foo) {
  case 1: {
    let x = 1;
    break;
  }
  case 2: {
    const y = 2;
    break;
  }
  case 3: {
    function f() {}
    break;
  }
  default: {
    class C {}
  }
}
```

## Original Documentation

https://eslint.org/docs/latest/rules/no-case-declarations
