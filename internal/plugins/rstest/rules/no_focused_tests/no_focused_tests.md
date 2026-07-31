# no-focused-tests

## Rule Details

Disallow focused Rstest tests and `describe` blocks. The rule reports an
explicit `.only` in a test or suite registration, including conditional,
parameterized, and extended test APIs.

The rule provides a suggestion that removes every `.only` from the
registration. If removing an accessor would also remove a comment, the rule
still reports it but does not provide a suggestion.

Examples of **incorrect** code for this rule:

```typescript
import { describe, test } from '@rstest/core';

test.only('adds two numbers', () => {});
describe.only('math', () => {});
test.concurrent.only('runs concurrently', () => {});
test.only.for([1, 2])('handles %s', () => {});
test.extend({ database }).only('uses a database', () => {});
```

Examples of **correct** code for this rule:

```typescript
import { describe, test } from '@rstest/core';

test('adds two numbers', () => {});
describe('math', () => {});
test.concurrent('runs concurrently', () => {});
test.extend({ database })('uses a database', () => {});
```

## References

- [Rstest `test` API](https://rstest.rs/api/runtime-api/test-api/test)
- [Rstest `describe` API](https://rstest.rs/api/runtime-api/test-api/describe)
