# prefer-hooks-in-order

## Rule Details

Prefer Jest lifecycle hooks in the same order Jest runs them.

Hooks can be written in any order, but Jest always executes them as:

1. `beforeAll`
2. `beforeEach`
3. `afterEach`
4. `afterAll`

This rule reports when a hook appears **after** a hook from a later stage in that
sequence within the same **consecutive group** of hook calls. Matching source order to
runtime order makes suite setup easier to read.

Examples of **incorrect** code for this rule:

```javascript
describe('foo', () => {
  beforeEach(() => {
    seedMyDatabase();
  });

  beforeAll(() => {
    createMyDatabase();
  });

  it('accepts this input', () => {
    // ...
  });

  it('returns that value', () => {
    // ...
  });

  describe('when the database has specific values', () => {
    const specificValue = '...';

    beforeEach(() => {
      seedMyDatabase(specificValue);
    });

    it('accepts that input', () => {
      // ...
    });

    it('throws an error', () => {
      // ...
    });

    afterEach(() => {
      clearLogger();
    });
    beforeEach(() => {
      mockLogger();
    });

    it('logs a message', () => {
      // ...
    });
  });

  afterAll(() => {
    removeMyDatabase();
  });
});
```

Examples of **correct** code for this rule:

```javascript
describe('foo', () => {
  beforeAll(() => {
    createMyDatabase();
  });

  beforeEach(() => {
    seedMyDatabase();
  });

  it('accepts this input', () => {
    // ...
  });

  it('returns that value', () => {
    // ...
  });

  describe('when the database has specific values', () => {
    const specificValue = '...';

    beforeEach(() => {
      seedMyDatabase(specificValue);
    });

    it('accepts that input', () => {
      // ...
    });

    it('throws an error', () => {
      // ...
    });

    beforeEach(() => {
      mockLogger();
    });

    afterEach(() => {
      clearLogger();
    });

    it('logs a message', () => {
      // ...
    });
  });

  afterAll(() => {
    removeMyDatabase();
  });
});
```

## Also See

- [`prefer-hooks-on-top`](/rules/jest/prefer-hooks-on-top) — require hooks before test cases in the same scope

## Original Documentation

- [jest/prefer-hooks-in-order](https://github.com/jest-community/eslint-plugin-jest/blob/main/docs/rules/prefer-hooks-in-order.md)
