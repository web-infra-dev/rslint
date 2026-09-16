# no-path-concat

Disallow string concatenation with `__dirname`, `__filename`, and `import.meta`
paths when the appended text starts with a path separator.

## Rule Details

Use `path.join()` or `path.resolve()` to build file paths. These functions handle
platform-specific separators and normalize the resulting path. Use `new URL()`
to resolve relative URLs against `import.meta.url`.

Examples of **incorrect** code:

```javascript
const config = __dirname + '/config.json';
const child = `${__filename}/child`;
const assets = import.meta.dirname + '/assets';
const sibling = import.meta.filename + '/sibling';
const url = `${import.meta.url}/assets`;
```

Examples of **correct** code:

```javascript
const config = path.join(__dirname, 'config.json');
const child = path.resolve(__filename, 'child');
const assets = path.join(import.meta.dirname, 'assets');
const sibling = path.resolve(import.meta.filename, 'sibling');
const url = new URL('./assets', import.meta.url);

const sourceMap = __filename + '.map';
const backup = `${import.meta.filename}.bak`;
```

The rule checks `+` expressions and template strings. It recognizes `/`, the
separator for the platform running the linter, and tracked `path.sep` reads.
Appending an extension or another suffix without a leading separator is allowed.

`__dirname` and `__filename` must be enabled globals. Local declarations and
parameters with these names are ignored, as are globals assigned anywhere in
the file. `import.meta` paths do not require configured globals.

## Options

This rule has no options and provides no automatic fixes or suggestions.

## Differences from upstream

Some computed separators are not recognized. For example,
`__dirname + String.fromCodePoint(47)` is reported by upstream but not by rslint.
Using the equivalent literal `/` or `path.sep` makes this concatenation detectable.

## References

- [Upstream documentation](https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/docs/rules/no-path-concat.md)
- [Upstream source](https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/lib/rules/no-path-concat.js)
