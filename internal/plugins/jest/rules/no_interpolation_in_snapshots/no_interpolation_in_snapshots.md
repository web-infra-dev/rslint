# no-interpolation-in-snapshots

## Rule Details

Disallow string interpolation inside inline snapshots. Interpolation prevents Jest from updating snapshots; instead, overload dynamic properties with a matcher via [property matchers](https://jestjs.io/docs/snapshot-testing#property-matchers).

Examples of **incorrect** code for this rule:

```javascript
expect(something).toMatchInlineSnapshot(
  `Object {
    property: ${interpolated}
  }`,
);

expect(something).toMatchInlineSnapshot(
  { other: expect.any(Number) },
  `Object {
    other: Any<Number>,
    property: ${interpolated}
  }`,
);

expect(errorThrowingFunction).toThrowErrorMatchingInlineSnapshot(
  `${interpolated}`,
);
```

Examples of **correct** code for this rule:

```javascript
expect(something).toMatchInlineSnapshot();

expect(something).toMatchInlineSnapshot(
  `Object {
    property: 1
  }`,
);

expect(something).toMatchInlineSnapshot(
  { property: expect.any(Date) },
  `Object {
    property: Any<Date>
  }`,
);

expect(errorThrowingFunction).toThrowErrorMatchingInlineSnapshot();

expect(errorThrowingFunction).toThrowErrorMatchingInlineSnapshot(
  `Error Message`,
);
```

## Original Documentation

- [jest/no-interpolation-in-snapshots](https://github.com/jest-community/eslint-plugin-jest/blob/main/docs/rules/no-interpolation-in-snapshots.md)
