# no-focused-tests

## Rule Details

Disallow focused Rstest tests and `describe` blocks. Focusing a test with the
`.only` modifier makes the runner execute only the focused tests and silently
skip everything else, so a committed `.only` disables most of the suite in CI.

Unlike Jest, Rstest has no `fit` / `fdescribe` aliases, so this rule only
reports the `.only` modifier. Any chain ordering is covered, e.g.
`test.concurrent.only`, `test.only.for(...)`, and `describe.only`.

The rule provides a suggestion that removes the `.only` modifier.

Examples of **incorrect** code for this rule:

```typescript
import { describe, test } from '@rstest/core';

test.only('adds two numbers', () => {});
describe.only('math', () => {});
test.concurrent.only('runs concurrently', () => {});
test.only.for([1, 2])('handles %s', () => {});
```

Examples of **correct** code for this rule:

```typescript
import { describe, test } from '@rstest/core';

test('adds two numbers', () => {});
describe('math', () => {});
test.concurrent('runs concurrently', () => {});
```

## References

- [Rstest `test` API](https://rstest.rs/api/runtime-api/test-api/test)
- [Rstest `describe` API](https://rstest.rs/api/runtime-api/test-api/describe)
