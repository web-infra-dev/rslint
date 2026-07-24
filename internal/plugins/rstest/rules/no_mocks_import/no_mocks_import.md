# no-mocks-import

## Rule Details

When using `rs.mock`, tests should import from the original module path, not directly from a `__mocks__` directory. Directly importing a manual mock can create a separate module instance and make assertions behave unexpectedly.

This rule reports `import` declarations and `require()` calls whose module specifier contains a `__mocks__` path segment.

Examples of **incorrect** code for this rule:

```typescript
import thing from './__mocks__/thing';
require('./__mocks__/thing');
```

Examples of **correct** code for this rule:

```typescript
rs.mock('./thing');
import thing from './thing';
```
