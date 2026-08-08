# no-interpolation-in-snapshots

## Rule Details

Disallow string interpolation inside inline snapshots. The expected value of an inline snapshot lives in the source file, and updating snapshots rewrites that literal — so the interpolation is evaluated away and silently lost. Overload dynamic properties with a property matcher instead.

Examples of **incorrect** code for this rule:

```javascript
import { expect } from '@rstest/core';

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
import { expect } from '@rstest/core';

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

## Only inline snapshots are checked

The rule covers `toMatchInlineSnapshot` and `toThrowErrorMatchingInlineSnapshot` only. Every other snapshot matcher keeps its expected value outside the source file, so nothing rewrites its arguments and interpolation is harmless:

- `toMatchSnapshot` and its Chai-style alias `matchSnapshot`
- `toThrowErrorMatchingSnapshot`
- `toMatchFileSnapshot`

`toMatchFileSnapshot` deserves a special mention: it has no Jest counterpart, and its first argument is a **file path**, so interpolating it is the intended usage rather than a mistake.

```javascript
// correct — the argument is a path, not a snapshot
await expect(data).toMatchFileSnapshot(`./__snapshots__/${name}.json`);
```

## Original Documentation

- [jest/no-interpolation-in-snapshots](https://github.com/jest-community/eslint-plugin-jest/blob/main/docs/rules/no-interpolation-in-snapshots.md)
