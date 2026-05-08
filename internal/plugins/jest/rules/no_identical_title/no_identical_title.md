# jest/no-identical-title

## Rule Details

Disallow the same title for two tests or two `describe` blocks **in the same scope**. Only **static** first arguments count (plain strings / simple templates); dynamic titles and `*.each` calls are ignored.

Examples of **incorrect** code for this rule:

```javascript
describe("foo", () => {
  it("bar", () => {});
  it("bar", () => {});
});
describe("x", () => {});
describe("x", () => {});
```

Examples of **correct** code for this rule:

```javascript
describe("foo", () => {
  it("a", () => {});
  it("b", () => {});
  describe("foo", () => {}); // same as parent name, different scope
});
test("x" + n, () => {});
test("x" + n, () => {}); // not static — skipped
```

## Original Documentation

- [jest/no-identical-title](https://github.com/jest-community/eslint-plugin-jest/blob/main/docs/rules/no-identical-title.md)
