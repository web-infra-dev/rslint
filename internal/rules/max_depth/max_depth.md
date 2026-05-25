# max-depth

## Rule Details

This rule enforces a maximum depth that blocks can be nested to reduce code complexity.

Examples of **incorrect** code for this rule with the default `{ "max": 4 }` option:

```javascript
function foo() {
  for (;;) {
    // Depth 1
    while (true) {
      // Depth 2
      if (true) {
        // Depth 3
        if (true) {
          // Depth 4
          if (true) {
            // Depth 5
          }
        }
      }
    }
  }
}
```

Examples of **correct** code for this rule with the default `{ "max": 4 }` option:

```javascript
function foo() {
  for (;;) {
    // Depth 1
    while (true) {
      // Depth 2
      if (true) {
        // Depth 3
        if (true) {
          // Depth 4
        }
      }
    }
  }
}
```

## Options

This rule accepts a number, or an object with `max` (default: `4`). The legacy
`maximum` key is also accepted for backward compatibility.

Examples of **incorrect** code for this rule with `{ "max": 2 }`:

```json
{ "max-depth": ["error", { "max": 2 }] }
```

```javascript
function foo() {
  if (true) {
    if (false) {
      if (true) {
      }
    }
  }
}
```

Examples of **correct** code for this rule with `{ "max": 2 }`:

```json
{ "max-depth": ["error", { "max": 2 }] }
```

```javascript
function foo() {
  if (true) {
    if (false) {
    }
  }
}
```

`else if` chains are treated as a single depth level. Class static blocks reset
the nesting counter — code inside `static {}` is measured from depth 0.

```javascript
function foo() {
  if (true) {
    // Depth 1
    class C {
      static {
        if (true) {
          // Depth 1 (resets in static block)
          if (true) {
            // Depth 2
          }
        }
      }
    }
  }
}
```

## Original Documentation

- [https://eslint.org/docs/latest/rules/max-depth](https://eslint.org/docs/latest/rules/max-depth)
