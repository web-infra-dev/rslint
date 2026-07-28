# require-hook

## Rule Details

Require setup and teardown code to be within a Jest lifecycle hook (`beforeEach`, `afterEach`, `beforeAll`, `afterAll`).

It's common when writing tests to need setup work before tests run and cleanup afterward. Because Jest executes all `describe` handlers in a test file **before** any of the actual tests, setup and teardown done directly in a `describe` body (or at the top level of the file) can run at the wrong time or leak across suites. This rule pushes that work into the lifecycle hooks instead.

This rule checks direct statements at the top level of a test file and direct statements within the body of each `describe` callback. It does not inspect arbitrary statements nested inside control-flow statements or other function callbacks.

It flags expressions in those scopes **except** for:

- `import` statements
- `const` variable declarations, including declarations with initializers
- Uninitialized `let` and `var` declarations
- `let` and `var` declarations initialized directly with `null` or `undefined`
- Classes
- Types
- Calls to recognized Jest APIs, such as `describe`, `test` / `it`, lifecycle hooks, `expect`, and `jest.*`

For variable declarations, a declaration containing any other initializer is reported as a whole. For example:

```javascript
let spy;                 // valid
let empty = null;        // valid
let notEmpty = create(); // incorrect
const value = create();  // valid
```

Examples of **incorrect** code for this rule:

```javascript
import { database, isCity } from '../database';
import { loadCities } from '../api';

jest.mock('../api');

const initializeCityDatabase = () => {
  database.addCity('Vienna');
  database.addCity('San Juan');
  database.addCity('Wellington');
};

const clearCityDatabase = () => {
  database.clear();
};

initializeCityDatabase();

test('that persists cities', () => {
  expect(database.cities.length).toHaveLength(3);
});

test('city database has Vienna', () => {
  expect(isCity('Vienna')).toBeTruthy();
});

test('city database has San Juan', () => {
  expect(isCity('San Juan')).toBeTruthy();
});

describe('when loading cities from the api', () => {
  let consoleWarnSpy = jest.spyOn(console, 'warn');

  loadCities.mockResolvedValue(['Wellington', 'London']);

  it('does not duplicate cities', async () => {
    await database.loadCities();

    expect(database.cities).toHaveLength(4);
  });

  it('logs any duplicates', async () => {
    await database.loadCities();

    expect(consoleWarnSpy).toHaveBeenCalledWith(
      'Ignored duplicate cities: Wellington',
    );
  });
});

clearCityDatabase();
```

Examples of **correct** code for this rule:

```javascript
import { database, isCity } from '../database';
import { loadCities } from '../api';

jest.mock('../api');

const initializeCityDatabase = () => {
  database.addCity('Vienna');
  database.addCity('San Juan');
  database.addCity('Wellington');
};

const clearCityDatabase = () => {
  database.clear();
};

beforeEach(() => {
  initializeCityDatabase();
});

test('that persists cities', () => {
  expect(database.cities.length).toHaveLength(3);
});

test('city database has Vienna', () => {
  expect(isCity('Vienna')).toBeTruthy();
});

test('city database has San Juan', () => {
  expect(isCity('San Juan')).toBeTruthy();
});

describe('when loading cities from the api', () => {
  let consoleWarnSpy;

  beforeEach(() => {
    consoleWarnSpy = jest.spyOn(console, 'warn');
    loadCities.mockResolvedValue(['Wellington', 'London']);
  });

  it('does not duplicate cities', async () => {
    await database.loadCities();

    expect(database.cities).toHaveLength(4);
  });

  it('logs any duplicates', async () => {
    await database.loadCities();

    expect(consoleWarnSpy).toHaveBeenCalledWith(
      'Ignored duplicate cities: Wellington',
    );
  });
});

afterEach(() => {
  clearCityDatabase();
});
```

## Options

- First argument (optional): object with `allowedFunctionCalls`
  - `allowedFunctionCalls`: array of exact callee names that are allowed outside hooks. Names are matched against the full dotted callee name, so `helper.setup` matches `helper.setup()` but `setup` does not.

```json
{ "jest/require-hook": ["error", { "allowedFunctionCalls": ["enableAutoDestroy"] }] }
```

```javascript
import { enableAutoDestroy } from '@vue/test-utils';
import { initDatabase, tearDownDatabase } from './databaseUtils';

enableAutoDestroy(afterEach);

beforeEach(initDatabase);
afterEach(tearDownDatabase);

describe('Foo', () => {
  test('always returns 42', () => {
    expect(global.getAnswer()).toBe(42);
  });
});
```

## Original Documentation

- [jest/require-hook](https://github.com/jest-community/eslint-plugin-jest/blob/main/docs/rules/require-hook.md)
