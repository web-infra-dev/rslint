# empty-brace-spaces

## Rule Details

Enforce no spaces between braces.

Empty blocks and object literals do not need internal whitespace, so this rule
enforces a compact, consistent style. It applies to block statements, class
bodies, static blocks, and object literals whose braces hold nothing but
whitespace.

Examples of **incorrect** code for this rule:

```javascript
class Unicorn {
}

try {
	foo();
} catch { }

const object = { };
```

Examples of **correct** code for this rule:

```javascript
class Unicorn {}

try {
	foo();
} catch {}

const object = {};
```

The rule has no options. It is automatically fixable by removing the
whitespace between the braces.

A comment between the braces counts as content, so the whitespace around it is
left untouched:

```javascript
class Unicorn { /* comment */ }
```

A brace pair in a class header is not mistaken for the class body: `class Foo
extends mixin({}) { }` reports the body's space and fixes it to `class Foo
extends mixin({}) {}`, and a header pair that is itself empty is reported
separately.

## Differences from ESLint

None.

## Original Documentation

- [eslint-plugin-unicorn: empty-brace-spaces](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v73.0.0/docs/rules/empty-brace-spaces.md)
- [Source code](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v73.0.0/rules/empty-brace-spaces.js)
