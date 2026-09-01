# require-unicode-regexp

## Rule Details

Enforces the use of the `u` or `v` flag on regular expressions.

Examples of **incorrect** code for this rule:

```javascript
const a = /aaa/;
const b = /bbb/gi;
const c = new RegExp("ccc");
const d = new RegExp("ddd", "gi");
```

Examples of **correct** code for this rule:

```javascript
const a = /aaa/u;
const b = /bbb/giu;
const c = new RegExp("ccc", "u");
const d = new RegExp("ddd", "giu");

const e = /aaa/v;
const f = /bbb/giv;
const g = new RegExp("ccc", "v");
const h = new RegExp("ddd", "giv");

// This rule ignores RegExp calls if the flags could not be evaluated to a static value.
function i(flags) {
  return new RegExp("eee", flags);
}
```

## Options

This rule has one object option:

- `requireFlag`: `"u"` or `"v"` — requires that particular flag instead of accepting either.

Examples of **incorrect** code for this rule with `{ "requireFlag": "u" }`:

```json
{ "require-unicode-regexp": ["error", { "requireFlag": "u" }] }
```

```javascript
const foo = /foo/;
const fooRegexp = new RegExp("foo");
const bar = /bar/v;
const barRegexp = new RegExp("bar", "v");
```

Examples of **correct** code for this rule with `{ "requireFlag": "u" }`:

```json
{ "require-unicode-regexp": ["error", { "requireFlag": "u" }] }
```

```javascript
const foo = /foo/u;
const fooRegexp = new RegExp("foo", "u");
```

Examples of **incorrect** code for this rule with `{ "requireFlag": "v" }`:

```json
{ "require-unicode-regexp": ["error", { "requireFlag": "v" }] }
```

```javascript
const foo = /foo/;
const fooRegexp = new RegExp("foo");
const bar = /bar/u;
const barRegexp = new RegExp("bar", "u");
```

Examples of **correct** code for this rule with `{ "requireFlag": "v" }`:

```json
{ "require-unicode-regexp": ["error", { "requireFlag": "v" }] }
```

```javascript
const foo = /foo/v;
const fooRegexp = new RegExp("foo", "v");
```

## Differences from ESLint

- Suggestions are conservatively omitted for every duplicate named capture,
  including the mutually exclusive alternatives that became valid in ES2025.
  The diagnostic is still reported, and an added flag never depends on
  duplicate-name control-flow acceptance.
- When validating a `v` suggestion, a negated character class containing a
  `\q{...}`, `\p{...}`, or `\P{...}` operand is conservatively treated as
  unsafe, including through nested sets. This can omit suggestions for
  single-code-point operands that ESLint proves safe, while preventing an
  invalid suggestion for string-valued operands.
- Static evaluation models RegExp literals and their stable properties, but it
  does not construct a new RegExp object while folding an expression. For
  example, `RegExp("g", "u").source` remains unknown even though ESLint can
  fold it to `"g"`.
## Original Documentation

- [ESLint: require-unicode-regexp](https://eslint.org/docs/latest/rules/require-unicode-regexp)
- [Source code](https://github.com/eslint/eslint/blob/v10.8.1/lib/rules/require-unicode-regexp.js)
