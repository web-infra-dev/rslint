# prefer-rs-mocked

## Rule Details

This rule requires `rs.mocked()` instead of mock type assertions. Type assertions such as `as Mock`, `as Mocked<T>`, and `as MockInstance` discard the original function signature, while `mocked()` adds mock methods without losing its parameter and return types.

The rule recognizes these types when they are imported from `@rstest/core` or `rstack/test`, including renamed imports and type-only imports. Other types with the same names, namespace-qualified types, `satisfies` expressions, and variable type annotations are not reported.

## Incorrect

```ts
import type { Mock } from '@rstest/core';

(loadUser as Mock).mockReturnValue({ id: 1, name: 'Ada' });
```

## Correct

```ts
import { rs } from '@rstest/core';

rs.mocked(loadUser).mockReturnValue({ id: 1, name: 'Ada' });
```

## Autofix

The autofix replaces the complete assertion with `rs.mocked(expression)` or `rstest.mocked(expression)`. It uses an imported utilities namespace when present, written under the name the import binds, so an aliased import is used under its alias. A type-only import binds no value and is not used. Otherwise, both global names must be free; the fix follows the spelling already used by the file and defaults to `rs`. No autofix is offered when an unrelated declaration makes the available global namespace uncertain.

## Suggestions

When no imported or global utilities namespace can be chosen safely but the file already uses `import.meta.rstest`, the rule suggests replacing the assertion with `import.meta.rstest.rs.mocked(expression)`. If none of these references can be used safely, the assertion is reported without an edit.
