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

- A pattern is judged against `u`-flag syntax when deciding whether the
  suggestion is safe, so one written in `v`-flag set notation — the difference
  operator (`[\d--[3]]`), a `\q{...}` string literal — is reported without a
  suggestion that ESLint would offer.
- Unicode property escapes are recognized by their short aliases, so
  `/\p{L}/` is offered the flag while `/\p{Letter}/` and
  `/\p{Script=Greek}/` are reported without a suggestion.
- A suggested flag insertion into a template literal that interpolates an
  expression is refused whenever that template's cooked value already
  contains the opposite flag character, even when the interpolated
  expression's source text has no escape sequence of its own — ESLint only
  inspects the template's literal quasis for that escape check, this rule
  inspects the whole template source.

## Original Documentation

- [ESLint: require-unicode-regexp](https://eslint.org/docs/latest/rules/require-unicode-regexp)
- [Source code](https://github.com/eslint/eslint/blob/v10.8.1/lib/rules/require-unicode-regexp.js)
