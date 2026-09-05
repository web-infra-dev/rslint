# padding-around-before-each-blocks

## Rule Details

Require a blank line before and after `beforeEach` statements so per-test setup is visually separated from tests and other declarations. No trailing blank line is required when the hook is the last statement in its scope.

The rule classifies an expression statement by its first identifier, so calls and modifier chains beginning with `beforeEach` are included regardless of where that name was declared. Renamed aliases and namespace calls such as `rstest.beforeEach` are not recognized. Type information is not required.

## Incorrect

```ts
const database = createDatabase();
beforeEach(() => database.reset());
test('loads a user', loadUser);
```

## Correct

```ts
const database = createDatabase();

beforeEach(() => database.reset());

test('loads a user', loadUser);
```

## Autofix

The rule inserts missing blank lines before and after `beforeEach` statements.
