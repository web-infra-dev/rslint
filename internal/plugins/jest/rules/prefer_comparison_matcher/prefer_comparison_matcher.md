# prefer-comparison-matcher

## Rule Details

Prefer Jest's built-in comparison matchers over wrapping a relational operator in
`expect(...).toBe(true)` (or `toEqual` / `toStrictEqual`).

Assertions such as `expect(x > 5).toBe(true)` are harder to read and produce less
helpful failure output than `expect(x).toBeGreaterThan(5)`.

This rule reports `expect(left OP right).<equalityMatcher>(true|false)` patterns
that can use one of these matchers instead:

- `toBeGreaterThan`
- `toBeGreaterThanOrEqual`
- `toBeLessThan`
- `toBeLessThanOrEqual`

Violations are automatically fixed where possible.

Examples of **incorrect** code for this rule:

```js
expect(x > 5).toBe(true);
expect(x < 7).not.toEqual(true);
expect(x <= y).toStrictEqual(true);
```

Examples of **correct** code for this rule:

```js
expect(x).toBeGreaterThan(5);
expect(x).not.toBeLessThanOrEqual(7);
expect(x).toBeLessThanOrEqual(y);

// special case - see below
expect(x < 'Carl').toBe(true);
```

**String comparisons.** These matchers only accept numbers and bigints. The rule
assumes operands are numeric and does not report comparisons that involve string
literals (for example, `expect(x < 'Carl').toBe(true)`). If you intentionally
compare strings with `>` or `<`, disable the rule for that line—otherwise the
fix rewrites the assertion to a numeric matcher and fails at runtime:

```js
// rslint-disable-next-line jest/prefer-comparison-matcher
expect(myName > theirName).toBe(true);
```

Negative expectations use the opposite comparison operator, matching eslint-plugin-jest. This assumes neither operand is `NaN`: for example, `expect(NaN > 1).toBe(false)` passes, but its fix `expect(NaN).toBeLessThanOrEqual(1)` fails. Disable the rule for comparisons that may involve `NaN` and need to retain that behavior.

## Differences from ESLint

- Parentheses around comma-expression operands are preserved, so moving `(read(), value)` does not turn it into multiple arguments.
- Type assertions on boolean expectations are removed with the boolean, rather than being applied to the replacement operand. For example, `toBe(true as const)` does not become `toBeGreaterThan(limit as const)`.
- String literals inside TypeScript type assertions are excluded just like unwrapped string literals.
- Only one diagnostic is emitted for the equality matcher when further calls follow it, such as `expect(a > b).toBe(true).toString()`.
- Assertions with parentheses around the receiver, such as `(expect(a > b)).toBe(true)`, are reported without an autofix to avoid removing only the closing parentheses.

## Original Documentation

- [eslint-plugin-jest: prefer-comparison-matcher](https://github.com/jest-community/eslint-plugin-jest/blob/v29.16.1/docs/rules/prefer-comparison-matcher.md)
- [Source code](https://github.com/jest-community/eslint-plugin-jest/blob/v29.16.1/src/rules/prefer-comparison-matcher.ts)
