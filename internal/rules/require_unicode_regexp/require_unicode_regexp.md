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

- When flags come from a newly constructed RegExp object's property, such as
  `RegExp("g", "u").source`, rslint may skip a call that ESLint reports.
- With computed `['__proto__']` properties in constant objects, rslint follows
  own-property semantics and can report a constructor call that ESLint skips.
- Capture names or Unicode properties newer than the bundled TypeScript
  parser's Unicode data can receive a diagnostic without a flag suggestion.
- For a negated `v` class where a range is followed by a string-valued operand,
  rslint omits ESLint's suggestion because JavaScript would reject the result.

## Original Documentation

- [ESLint: require-unicode-regexp](https://eslint.org/docs/latest/rules/require-unicode-regexp)
- [Source code](https://github.com/eslint/eslint/blob/v10.8.1/lib/rules/require-unicode-regexp.js)
