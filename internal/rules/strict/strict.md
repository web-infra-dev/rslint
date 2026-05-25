# strict

## Rule Details

Require or disallow strict mode directives (`"use strict"`).

The rule supports four options:

- `"safe"` (default) — equivalent to `"function"` for script files; module files always use `"module"` semantics.
- `"never"` — disallows all strict mode directives.
- `"global"` — requires exactly one strict directive in global scope and disallows all other directives.
- `"function"` — requires one strict directive in each top-level function and disallows directives in the global scope or in nested functions / class bodies.

When the file is an ES module (detected via top-level `import` / `export`), the rule always uses module semantics: every `"use strict"` directive is reported as unnecessary and removed by autofix.

## Examples

### `"never"`

```json
{ "strict": ["error", "never"] }
```

Examples of **incorrect** code:

```javascript
"use strict";
function foo() {}
```

```javascript
function foo() {
    "use strict";
}
```

Examples of **correct** code:

```javascript
function foo() {}
```

### `"global"`

```json
{ "strict": ["error", "global"] }
```

Examples of **incorrect** code:

```javascript
function foo() {}
```

```javascript
function foo() {
    "use strict";
}
```

```javascript
"use strict";
function foo() {
    "use strict";
}
```

Examples of **correct** code:

```javascript
"use strict";
function foo() {}
```

### `"function"`

```json
{ "strict": ["error", "function"] }
```

Examples of **incorrect** code:

```javascript
"use strict";
function foo() {}
```

```javascript
function foo() {}
(function() {
    function bar() {
        "use strict";
    }
}());
```

Examples of **correct** code:

```javascript
function foo() {
    "use strict";
}

(function() {
    "use strict";
    function bar() {}
    function baz(a = 1) {}
}());

const foo2 = (function() {
    "use strict";
    return function foo(a = 1) {};
}());
```

## Original Documentation

- https://eslint.org/docs/latest/rules/strict
