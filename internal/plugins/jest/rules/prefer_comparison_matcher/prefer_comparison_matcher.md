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

## Original Documentation

- [jest/prefer-comparison-matcher](https://github.com/jest-community/eslint-plugin-jest/blob/main/docs/rules/prefer-comparison-matcher.md)
