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
- TypeScript-only wrappers with no runtime effect — `x!`, `x as T`, `x satisfies T` — are treated the same as a bare receiver when deciding whether the two sides name the same value. For example, `x!.y = x!.y + z` is autofixed to `x!.y += z`.
- The shorthand form drops the right-hand copy of the target, so a fix is offered only when both copies carry the same assertions. `x = (x as number) * 2` and `x = x! + 1` are reported without a fix, since `x *= 2` would type-check `x` against its declared type again.
- In `"never"` mode, a right side whose leftmost operand is a bare `as` or `satisfies` expression is parenthesized as a whole: `x += a as number * b` becomes `x = x + (a as number * b)`, which keeps the original grouping. Without the parentheses TypeScript would read the result as `((x + a) as number) * b`.
- In `"never"` mode, `<<=` is reported without a fix when the right side both needs parentheses and writes TypeScript type syntax of its own — a type argument list such as `x <<= foo<T>` or a type parameter list such as `x <<= <X>(v: X) => v`. TypeScript re-scans the `<` of `<<` as the start of a type argument list, and the right side's own `>` closes it, so `x = x << (foo<T>)` would not parse. Angle brackets that are not type syntax — `x <<= a < b`, a `<` inside a string, template, or comment — are parenthesized and fixed as usual.

## Original Documentation

- [ESLint: operator-assignment](https://eslint.org/docs/latest/rules/operator-assignment)
- [Source code](https://github.com/eslint/eslint/blob/v10.8.1/lib/rules/operator-assignment.js)
