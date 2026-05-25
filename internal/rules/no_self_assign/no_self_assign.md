# no-self-assign

## Rule Details

Disallow assignments where both sides are exactly the same. Self-assignments have no effect, so they are probably errors due to incomplete refactoring.

Examples of **incorrect** code for this rule:

```javascript
a = a;
[a, b] = [a, b];
[a, ...b] = [a, ...b];
({ a } = { a });
({ a: b } = { a: b });
a.b = a.b;
a.b.c = a.b.c;
a[0] = a[0];
a.b = a['b'];
a &&= a;
a ||= a;
a ??= a;
```

Examples of **correct** code for this rule:

```javascript
a = b;
[a, b] = [b, a];
a.b = a.c;
a.b = c.b;
a += a;
a = +a;
a.b = a?.b; // considered self-assignment
```

## Options

This rule has an object option:

- `props` (boolean, default: `true`): When `true`, checks member expression (property access and element access) self-assignments such as `a.b = a.b` and `a[0] = a[0]`. Set to `false` to disable property checks.

Examples of **correct** code with `{ "props": false }`:

```javascript
a.b = a.b;
a[0] = a[0];
this.x = this.x;
```

## Original Documentation

https://eslint.org/docs/latest/rules/no-self-assign
