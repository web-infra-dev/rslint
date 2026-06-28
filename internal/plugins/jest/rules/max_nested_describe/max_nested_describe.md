# max-nested-describe

## Rule Details

Enforce a maximum depth for nested `describe()` calls. Grouping tests with `describe` is useful, but too many nested levels make suites harder to read and navigate.

This rule counts every Jest suite call as a nesting level, including `fdescribe`, `xdescribe`, `describe.only`, `describe.skip`, and `describe.each`.

Examples of **incorrect** code for this rule (with the default `{ "max": 5 }`):

```javascript
describe('foo', () => {
  describe('bar', () => {
    describe('baz', () => {
      describe('qux', () => {
        describe('quxx', () => {
          describe('too many', () => {
            it('should get something', () => {
              expect(getSomething()).toBe('Something');
            });
          });
        });
      });
    });
  });
});

describe('foo', function () {
  describe('bar', function () {
    describe('baz', function () {
      describe('qux', function () {
        describe('quxx', function () {
          describe('too many', function () {
            it('should get something', () => {
              expect(getSomething()).toBe('Something');
            });
          });
        });
      });
    });
  });
});
```

Examples of **correct** code for this rule (with the default `{ "max": 5 }`):

```javascript
describe('foo', () => {
  describe('bar', () => {
    it('should get something', () => {
      expect(getSomething()).toBe('Something');
    });
  });

  describe('qux', () => {
    it('should get something', () => {
      expect(getSomething()).toBe('Something');
    });
  });
});

describe('foo2', function () {
  it('should get something', () => {
    expect(getSomething()).toBe('Something');
  });
});

describe('foo', function () {
  describe('bar', function () {
    describe('baz', function () {
      describe('qux', function () {
        describe('this is the limit', function () {
          it('should get something', () => {
            expect(getSomething()).toBe('Something');
          });
        });
      });
    });
  });
});
```

## Options

- First argument (optional): object with `max`
  - `max`: maximum allowed nesting depth for `describe()` calls. Default is `5`. A value of `0` disallows any `describe` block.

Examples of **correct** code with `{ "max": 2 }`:

```javascript
describe('foo', () => {
  describe('bar', () => {
    it('should get something', () => {
      expect(getSomething()).toBe('Something');
    });
  });
});
```

Examples of **incorrect** code with `{ "max": 2 }`:

```javascript
fdescribe('foo', () => {
  describe.only('bar', () => {
    describe.skip('baz', () => {
      it('should get something', () => {
        expect(getSomething()).toBe('Something');
      });
    });
  });
});
```

## Original Documentation

- [jest/max-nested-describe](https://github.com/jest-community/eslint-plugin-jest/blob/main/docs/rules/max-nested-describe.md)
