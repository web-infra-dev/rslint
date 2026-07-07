# no-cycle

## Rule Details

Ensures that an imported module does not have a resolvable dependency path back
to the linted module.

Examples of **incorrect** code for this rule:

```javascript
// dep-b.js
import "./dep-a.js";

// dep-a.js
import "./dep-b.js";
```

Examples of **correct** code for this rule:

```javascript
// dep-b.js
export const value = 1;

// dep-a.js
import { value } from "./dep-b.js";
```

This rule does not report direct self imports. Use `import/no-self-import` for
that case.

Type-only imports are ignored because they have no runtime effect. Named
type-only re-exports still participate in the dependency graph, matching
`eslint-plugin-import`.

## Options

### `maxDepth`

Limits how far the rule traverses the dependency graph. The value must be a
positive integer or `"∞"`.

```json
{ "import/no-cycle": ["error", { "maxDepth": 1 }] }
```

### `commonjs`

Checks `require()` calls in addition to ES module imports.

```json
{ "import/no-cycle": ["error", { "commonjs": true }] }
```

### `amd`

Checks AMD `require([...])` and `define([...])` dependencies.

```json
{ "import/no-cycle": ["error", { "amd": true }] }
```

### `ignoreExternal`

Skips modules treated as external, such as modules under `node_modules`.

```json
{ "import/no-cycle": ["error", { "ignoreExternal": true }] }
```

### `allowUnsafeDynamicCyclicDependency`

Allows a cycle when at least one dependency in the cycle is imported with
dynamic `import()`.

```json
{
  "import/no-cycle": [
    "error",
    { "allowUnsafeDynamicCyclicDependency": true }
  ]
}
```

## Original Documentation

- [eslint-plugin-import/no-cycle](https://github.com/import-js/eslint-plugin-import/blob/main/docs/rules/no-cycle.md)
