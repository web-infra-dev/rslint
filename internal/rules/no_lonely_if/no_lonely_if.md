# no-lonely-if

## Rule Details

If an `if` statement is the only statement in the `else` block, it is often clearer to use an `else if` form.

```js
if (foo) {
    // ...
} else {
    if (bar) {
        // ...
    }
}
```

should be rewritten as

```js
if (foo) {
    // ...
} else if (bar) {
    // ...
}
```

This rule disallows `if` statements as the only statement in `else` blocks.

Examples of **incorrect** code for this rule:

```javascript
if (condition) {
    // ...
} else {
    if (anotherCondition) {
        // ...
    }
}

if (condition) {
    // ...
} else {
    if (anotherCondition) {
        // ...
    } else {
        // ...
    }
}
```

Examples of **correct** code for this rule:

```javascript
if (condition) {
    // ...
} else if (anotherCondition) {
    // ...
}

if (condition) {
    // ...
} else if (anotherCondition) {
    // ...
} else {
    // ...
}

if (condition) {
    // ...
} else {
    if (anotherCondition) {
        // ...
    }
    doSomething();
}
```

## When Not To Use It

Disable this rule if the code is clearer without requiring the `else if` form.

## Original Documentation

- [ESLint: no-lonely-if](https://eslint.org/docs/latest/rules/no-lonely-if)
- [Source code](https://github.com/eslint/eslint/blob/v10.8.1/lib/rules/no-lonely-if.js)
