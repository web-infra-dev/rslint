# no-self-assign

## Rule Details

Disallow assignments where both sides are exactly the same. Self-assignments have no effect, so they are probably errors due to incomplete refactoring.

Examples of **incorrect** code for this rule:

```javascript
a = a;
[a, b] = [a, b];
a.b = a.b;
a[0] = a[0];
```

Examples of **correct** code for this rule:

```javascript
a = b;
[a, b] = [b, a];
a.b = a.c;
a += a;
```

## Options

- `props`: If `true` (default), checks member expression self-assignment such as `a.b = a.b`. Set to `false` to disable property checks.

## Original Documentation

https://eslint.org/docs/latest/rules/no-self-assign
