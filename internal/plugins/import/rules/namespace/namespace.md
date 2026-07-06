# import/namespace

## Rule Details

Enforces that properties read from namespace imports exist in the imported
module.

Examples of **incorrect** code for this rule:

```javascript
import * as names from './named-exports';

names.missing;
```

```javascript
import * as names from './named-exports';

names['dynamic'];
```

```javascript
import * as names from './named-exports';

names.foo = 1;
```

Examples of **correct** code for this rule:

```javascript
import * as names from './named-exports';

names.foo;
```

Examples of **correct** code for this rule with `{ "allowComputed": true }`:

```json
{ "import/namespace": ["error", { "allowComputed": true }] }
```

```javascript
import * as names from './named-exports';

names[key];
```

## Options

### `allowComputed`

Defaults to `false`. When set to `true`, computed namespace member access is
allowed, but the computed property name is not validated.

Modules that cannot be resolved, are ignored, or are not ES modules are not reported by this rule.

## Original Documentation

- [eslint-plugin-import/namespace](https://github.com/import-js/eslint-plugin-import/blob/main/docs/rules/namespace.md)
