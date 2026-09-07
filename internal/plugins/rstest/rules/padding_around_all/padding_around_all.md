# padding-around-all

## Rule Details

Require blank lines around Rstest lifecycle hooks, suites, tests, and assertion groups. This rule combines the behavior of the seven focused `padding-around-*` rules. Consecutive `expect` statements form one group and do not require blank lines between them.

The rule classifies an expression statement by its first identifier, so calls and modifier or parameterization chains beginning with `beforeAll`, `beforeEach`, `afterAll`, `afterEach`, `describe`, `test`, `it`, or `expect` are included regardless of where that name was declared. Renamed aliases and namespace calls such as `rstest.test` are not recognized. Type information is not required.

## Incorrect

```ts
const database = createDatabase();
beforeAll(() => database.connect());
test('loads a user', () => {
  const user = loadUser();
  expect(user.name).toBe('Ada');
});
```

## Correct

```ts
const database = createDatabase();

beforeAll(() => database.connect());

test('loads a user', () => {
  const user = loadUser();

  expect(user.name).toBe('Ada');
});
```

## Autofix

The rule inserts missing blank lines around matching statements and assertion groups.
