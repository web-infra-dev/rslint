# prefer-comparison-matcher

## Rule Details

Prefer numeric comparison matchers over checking a comparison's boolean result with `toBe`, `toEqual`, or `toStrictEqual`. For example, `expect(duration).toBeLessThan(timeout)` provides more useful failure messages than `expect(duration < timeout).toBe(true)`. The rule handles `>`, `>=`, `<`, and `<=`, literal boolean expectations, and `.not`. Negative expectations use `.not` with the original comparison matcher, rather than the opposite operator, because comparisons involving `NaN` are always false.

The rule recognizes Rstest globals, imports and import aliases from `@rstest/core` and `@rstest/playwright`, namespace and CommonJS access, `import.meta.rstest`, and the local `expect` from a test's [TestContext](https://rstest.rs/api/runtime-api/test-api/test#testcontext). Both `expect(value, message)` and `expect.soft(value, message)` are checked. Dot access and static string or template matcher names are recognized, including Chai language chains before these three equality matchers; Chai's own equality methods and property assertions are not checked.

The rule does not require type information. Comparisons with string or template literals, boolean or null literals, arrays, objects, functions, classes, regular expressions, or `void` operands are excluded, including through TypeScript assertions. Other operands are assumed to be numeric: use this rule for number and bigint comparisons, not comparisons relying on implicit conversion. Dynamic matcher names, shadowed or foreign `expect` functions, `expect.poll`, `expect.element`, and chains with `resolves` or `rejects` are excluded. A relational expression returns a boolean, not the promise those modifiers require. Only the first matcher in a chain is checked.

## Incorrect

```ts
test('keeps processing within the limit', () => {
  expect(processingTime < timeout).toBe(true);
  expect(retryCount > maximumRetries).toEqual(false);
});
```

## Correct

```ts
test('keeps processing within the limit', () => {
  expect(processingTime).toBeLessThan(timeout);
  expect(retryCount).not.toBeGreaterThan(maximumRetries);
});
```

## Suggestions

When both operands are numeric or bigint constants, the suggestion moves the left operand into `expect`, moves the right operand into the comparison matcher, and adjusts `.not` according to the boolean expectation. It preserves accessor quoting, optional chaining, and trailing commas. The change is a suggestion rather than an autofix because numeric matchers reject values that JavaScript comparison operators implicitly convert.

Other operands are still reported without a suggestion. Calling `expect` before evaluating the right operand changes evaluation order, while passing an object to `expect` postpones numeric coercion until the matcher runs. Explicit type arguments on `expect` or the equality matcher suppress the suggestion because they describe the original boolean assertion. Comments inside a replaced comparison or boolean argument and members following the equality matcher suppress suggestions as well.
