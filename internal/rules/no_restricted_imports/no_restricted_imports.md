# no-restricted-imports

## Rule Details

Disallow specified modules when loaded by `import`. This rule allows you to specify imports that you don't want to use in your application.

This can be useful if you want to restrict usage of certain modules, enforce alternatives, or prevent accidental use of deprecated APIs.

Examples of **incorrect** code for this rule with `["error", "fs"]`:

```javascript
import fs from 'fs';
export * from 'fs';
```

Examples of **correct** code for this rule with `["error", "fs"]`:

```javascript
import crypto from 'crypto';
```

## Options

The rule accepts either an array of strings/objects or an object with `paths` and `patterns` properties.

### String format

```json
"no-restricted-imports": ["error", "fs", "path"]
```

### Object format with paths and patterns

```json
"no-restricted-imports": ["error", {
  "paths": [{
    "name": "import-foo",
    "importNames": ["Bar"],
    "message": "Please use Bar from /import-bar/ instead."
  }],
  "patterns": [{
    "group": ["import1/private/*"],
    "message": "usage of import1 private modules not allowed."
  }]
}]
```

### Path options

- `name` (required): The module name to restrict
- `message`: Custom message to display
- `importNames`: Restrict specific named exports
- `allowImportNames`: Allow only specified named exports
- `allowTypeImports`: Allow type-only imports (TypeScript)

### Pattern options

- `group`: Gitignore-style patterns
- `regex`: Regular expression pattern
- `message`: Custom message to display
- `caseSensitive`: Case-sensitive matching (default: false)
- `importNames`: Restrict specific named imports
- `importNamePattern`: Regex pattern for import names
- `allowImportNames`: Allow only specified named imports
- `allowImportNamePattern`: Regex pattern for allowed import names
- `allowTypeImports`: Allow type-only imports (TypeScript)

## Original Documentation

[ESLint: no-restricted-imports](https://eslint.org/docs/latest/rules/no-restricted-imports)
