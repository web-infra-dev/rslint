# prefer-hooks-on-top

## Rule Details

Suggest placing Jest lifecycle hooks before any test cases in the same scope.

Hooks (`beforeAll`, `beforeEach`, `afterEach`, `afterAll`) can appear anywhere in a
`describe` callback, but Jest always runs them in a fixed order regardless of where
they are written. Mixing hooks with `test` / `it` calls makes that execution order
harder to follow, so this rule reports hooks registered **after** the first test case
in the same scope.

Examples of **incorrect** code for this rule:

```javascript
describe('foo', () => {
  beforeEach(() => {
    seedMyDatabase();
  });

  it('accepts this input', () => {
    // ...
  });

  beforeAll(() => {
    createMyDatabase();
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

  afterAll(() => {
    clearMyDatabase();
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

    beforeEach(() => {
      mockLogger();
    });

    afterEach(() => {
      clearLogger();
    });

    it('accepts that input', () => {
      // ...
    });

    it('throws an error', () => {
      // ...
    });

    it('logs a message', () => {
      // ...
    });
  });
});
```

## Differences from upstream

Every function body is a scope of its own, where upstream only opens a new scope at a
call expression. A hook written in a function that is never called registers nothing at
run time, and a function passed to `describe` by name is that suite's body rather than a
continuation of the code around it, so neither inherits the test cases registered
outside it:

```javascript
describe('foo', () => {
  test('bar', () => {});

  // Not reported here; upstream reports it.
  function unused() {
    beforeEach(() => {});
  }
});
```

The other side of that boundary is that a hook registered through a helper the suite
does call is judged against the helper instead of the suite, and so goes unreported.

## Original Documentation

- [eslint-plugin-jest: prefer-hooks-on-top](https://github.com/jest-community/eslint-plugin-jest/blob/v29.16.0/docs/rules/prefer-hooks-on-top.md)
- [Source code](https://github.com/jest-community/eslint-plugin-jest/blob/v29.16.0/src/rules/prefer-hooks-on-top.ts)
