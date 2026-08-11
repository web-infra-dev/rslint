# no-empty-static-block

## Rule Details

Disallow empty class static blocks. Empty static blocks usually indicate that a
refactor was left unfinished or that an intentional placeholder needs an
explanatory comment.

Examples of **incorrect** code for this rule:

```javascript
class Foo {
  static {}
}

class Bar {
  static {
  }
}
```

Examples of **correct** code for this rule:

```javascript
class Foo {
  static {
    initialize();
  }
}

class Bar {
  static {
    // intentionally empty
  }
}
```

## Original Documentation

- [ESLint: no-empty-static-block](https://eslint.org/docs/latest/rules/no-empty-static-block)
- [Source code](https://github.com/eslint/eslint/blob/v10.8.1/lib/rules/no-empty-static-block.js)
