# operator-assignment

## Rule Details

This rule requires or disallows assignment operator shorthand where possible.

This rule applies to the following 12 operators: `+=`, `-=`, `*=`, `/=`, `%=`, `**=`, `<<=`, `>>=`, `>>>=`, `&=`, `^=`, `|=`. Logical assignment operators (`&&=`, `||=`, `??=`) are not checked, since they short-circuit differently from a plain assignment.

Examples of **incorrect** code for this rule, in the default `"always"` mode:

```javascript
x = x + y;
x = y * x;
x[0] = x[0] / y;
x.y = x.y << z;
```

Examples of **correct** code for this rule, in the default `"always"` mode:

```javascript
x = y;
x += y;
x = y * z;
x = (x * y) * z;
x[0] /= y;
x[foo()] = x[foo()] % 2;
x = y + x; // `+` is not always commutative
```

Examples of **incorrect** code for this rule with `"never"`:

```json
{ "operator-assignment": ["error", "never"] }
```

```javascript
x *= y;
x ^= (y + z) / foo();
```

Examples of **correct** code for this rule with `"never"`:

```json
{ "operator-assignment": ["error", "never"] }
```

```javascript
x = x + y;
x.y = x.y / a.b;
```

## Differences from ESLint

- Autofix is skipped when the assignment target is (or contains) an optional-chain access — for example, `obj.a = obj?.a + b`. ESLint also reports the diagnostic without a fix in this case.
- TypeScript-only wrappers with no runtime effect — `x!`, `x as T`, `x satisfies T` — are treated the same as a bare receiver for autofix purposes. For example, `x!.y = x!.y + z` is autofixed to `x!.y += z`.

## Original Documentation

- [ESLint: operator-assignment](https://eslint.org/docs/latest/rules/operator-assignment)
- [Source code](https://github.com/eslint/eslint/blob/v10.8.1/lib/rules/operator-assignment.js)
