# first

## Rule Details

Ensures all import statements appear before other statements in a module. Since imports are hoisted, interleaving them with other code can be confusing.

Examples of **incorrect** code for this rule:

```javascript
import { x } from './foo';
export { x };
import { y } from './bar';
```

```javascript
var a = 1;
import { y } from './bar';
```

Examples of **correct** code for this rule:

```javascript
import { x } from './foo';
import { y } from './bar';
export { x, y };
```

## Options

### `absolute-first`

When set to `"absolute-first"`, this rule enforces that absolute (package) imports appear before relative imports.

Examples of **incorrect** code with `"absolute-first"`:

```javascript
import { x } from './foo';
import { y } from 'bar';
```

Examples of **correct** code with `"absolute-first"`:

```javascript
import { y } from 'bar';
import { x } from './foo';
```

## Original Documentation

https://github.com/import-js/eslint-plugin-import/blob/main/docs/rules/first.md
