# prefer-query-selector

## Rule Details

Prefer `.querySelector()` and `.querySelectorAll()` over older DOM query methods.

This rule reports `.getElementById()`, `.getElementsByClassName()`,
`.getElementsByTagName()`, and `.getElementsByName()` calls to keep DOM queries
consistent. It automatically converts safe literal arguments to CSS selectors.
A following `.item(0)` becomes a single-element query when it is safe to remove
the access. Indexing with `[0]` and calls to `.at(0)` are preserved.

Examples of **incorrect** code:

```javascript
document.getElementById('foo');
document.getElementsByClassName('foo bar');
document.getElementsByTagName('main');
document.getElementsByName('email');
document.getElementsByTagName('form')[0];
element.getElementsByClassName('foo');
```

Examples of **correct** code:

```javascript
document.querySelector('#foo');
document.querySelectorAll('.foo.bar');
document.querySelectorAll('main');
document.querySelectorAll('[name="email"]');
document.querySelectorAll('form')[0];
element.querySelectorAll('.foo');
```

Values requiring CSS escaping, dynamic arguments, and empty strings are reported
without an automatic fix. Compose and escape dynamic selectors explicitly:

```javascript
const singleClassName = 'foo';

// Incorrect
document.getElementsByClassName(singleClassName);

// Correct
document.querySelectorAll(`.${CSS.escape(singleClassName)}`);
```

## Options

### `allowWithVariables`

Type: `boolean`\
Default: `false`

When `true`, allows non-literal arguments to `.getElementById()` and
`.getElementsByClassName()`. String literals, binary expressions, and template
literals containing non-whitespace text are still reported. Calls to
`.getElementsByTagName()` and `.getElementsByName()` are always checked.

```javascript
// With { allowWithVariables: true }
document.getElementById(someId);
document.getElementsByClassName(someClass);

// Still reported
document.getElementById('foo');
document.getElementsByClassName('foo');
document.getElementsByTagName(someTag);
document.getElementsByName(someName);
```

## Differences from upstream

The following differences from eslint-plugin-unicorn v79.0.0 keep autofixes from
changing query results:

- Generated selectors omit `:scope`. For example, `fragment.getElementById('foo')`
  becomes `fragment.querySelector('#foo')`, which also works on shadow roots.
- ID, class, and name queries with `null` arguments are reported without a fix.
  For example, `document.getElementById(null)` queries the ID `'null'`, whereas
  `document.querySelector(null)` queries a tag named `'null'`.
- Class names are split only on DOM ASCII whitespace. A query such as
  `document.getElementsByClassName('foo\u00a0bar')` is reported without a fix
  because the non-breaking space belongs to one class name.
- Name queries containing actual form feed or NUL characters are reported without
  a fix because those characters change meaning inside CSS attribute selectors.
- `[0]` and `.at(0)` are preserved because they return `undefined` for an empty
  collection. Only `.item(0)` can become `.querySelector()`, which also returns
  `null` when nothing matches.
- Class and tag queries whose result is directly accessed through `.namedItem`
  are reported without a fix because `NodeList` has no such member. This includes
  computed access such as `collection['namedItem']`.

Collection-returning fixes still produce a static `NodeList`, as upstream does.
Code that retains a live collection across DOM mutations needs manual review.

## Original Documentation

- [eslint-plugin-unicorn: prefer-query-selector](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/docs/rules/prefer-query-selector.md)
- [Source code](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/rules/prefer-query-selector.js)
