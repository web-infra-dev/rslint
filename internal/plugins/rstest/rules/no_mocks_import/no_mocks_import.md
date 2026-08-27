# no-mocks-import

## Rule Details

Disallows importing a manual mock directly from a `__mocks__` directory. Tests should mock the original module path so Rstest controls which module instance is used.

The rule checks both `import` declarations and `require()` calls.

## Incorrect

```ts
import userClient from './__mocks__/user-client';
```

## Correct

```ts
rs.mock('./user-client');
import userClient from './user-client';
```
