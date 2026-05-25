# newline-after-import

## Rule Details

Enforces having one or more empty lines after the last top-level import statement or require call.

This rule supports the following options:

- `count` which sets the number of newlines that are enforced after the last top-level import statement or require call. This option defaults to `1`.
- `exactCount` which enforces the exact number of newlines mentioned in `count`. This option defaults to `false`.
- `considerComments` which enforces the rule on comments after the last import statement as well when set to true. This option defaults to `false`.

Examples of **incorrect** code for this rule:

```javascript
import * as foo from 'foo';
const FOO = 'BAR';
```

```javascript
const FOO = require('./foo');
const BAZ = 1;
```

Examples of **correct** code for this rule:

```javascript
import defaultExport from './foo';

const FOO = 'BAR';
```

```javascript
const FOO = require('./foo');
const BAR = require('./bar');

const BAZ = 1;
```

## Original Documentation

- [eslint-plugin-import/newline-after-import](https://github.com/import-js/eslint-plugin-import/blob/main/docs/rules/newline-after-import.md)
