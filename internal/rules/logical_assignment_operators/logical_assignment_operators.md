# logical-assignment-operators

## Rule Details

This rule requires or disallows the logical assignment operators `&&=`, `||=`, and `??=`.

By default (`"always"`) it reports an assignment or a logical expression that duplicates its own target and can be written with a logical assignment operator instead.

Examples of **incorrect** code for this rule:

```javascript
a = a || b;
a = a && b;
a = a ?? b;

a || (a = b);
a && (a = b);
a ?? (a = b);
```

Examples of **correct** code for this rule:

```javascript
a ||= b;
a &&= b;
a ??= b;

a = b || c;
a || (b = c);
```

### enforceForIfStatements

With `{ "enforceForIfStatements": true }` the rule also reports an `if` statement whose only purpose is to assign to the value it tests.

```json
{ "logical-assignment-operators": ["error", "always", { "enforceForIfStatements": true }] }
```

Examples of **incorrect** code:

```javascript
if (a) a = b;
if (!a) a = b;
if (Boolean(a)) a = b;
if (a == null) a = b;
if (a === null || a === undefined) a = b;
```

Examples of **correct** code:

```javascript
if (a) b = c;
if (a) a = b; else a = c;
if (predicate(a)) a = b;
```

### never

With `"never"` the rule reports every logical assignment operator and prefers the expanded form.

```json
{ "logical-assignment-operators": ["error", "never"] }
```

Examples of **incorrect** code:

```javascript
a ||= b;
a &&= b;
a ??= b;
```

Examples of **correct** code:

```javascript
a = a || b;
a = a && b;
a = a ?? b;
```

## Fixes and suggestions

The rewrite is applied automatically when it cannot change how many times a getter or a setter runs, and is offered as a suggestion otherwise.

For `a = a || b` and for the `never` option, the target has to be a plain identifier. For `a || (a = b)` and for `if` statements, a single property access such as `a.b` also qualifies, as long as an `if` condition is a single test rather than a `||` chain. A bare name inside a non-strict `with` block may resolve to a property, so it counts as a single property access rather than a plain identifier.

A TypeScript assertion makes the two sides count as different values, so `a = a! || b` and `a[b!] || (a[b!] = c)` are left alone.

## Original Documentation

- [ESLint: logical-assignment-operators](https://eslint.org/docs/latest/rules/logical-assignment-operators)
- [Source code](https://github.com/eslint/eslint/blob/v10.8.1/lib/rules/logical-assignment-operators.js)
