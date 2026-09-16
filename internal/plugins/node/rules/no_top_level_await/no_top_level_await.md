# no-top-level-await

Disallow top-level `await` in published modules.

## Rule details

ES modules that use top-level `await` cannot be loaded with `require()`.
This rule helps libraries support both import styles by reporting top-level
`await`, `for await...of`, and `await using` in published files. Async functions
and their bodies are allowed.

A file is checked when its nearest `package.json` includes it in `files`, or
has no `files` field, and `.npmignore` does not exclude it. `.gitignore` is used
when both `files` and `.npmignore` are absent. The package's `main` entry remains
published. Files outside a package are ignored.

Examples of **incorrect** code in a published file:

```javascript
const data = await load();
for await (const item of stream) {
  consume(item);
}
```

Example of **correct** code:

```javascript
export async function loadData() {
  return await load();
}
```

This rule provides no automatic fixes or suggestions.

## Options

```javascript
export default [
  {
    plugins: ['node'],
    rules: {
      'node/no-top-level-await': ['error', {
        ignoreBin: false,
      }],
    },
  },
];
```

### ignoreBin

Defaults to `false`. Set it to `true` to allow top-level `await` in files listed
in `package.json`'s `bin` field or whose source starts with `#!/usr/bin/env`.

### convertPath

Map source paths to published paths before checking publication and `bin`.
For example, when `package.json` publishes `lib`, check matching TypeScript
sources with:

```javascript
{
  convertPath: [{
    include: ['src/**'],
    exclude: ['src/test/**'],
    replace: ['^src/(.*)\\.ts$', 'lib/$1.js'],
  }],
}
```

The first matching array entry wins. An object mapping patterns to
`[regularExpression, replacement]` pairs is also supported. Omit the option to
use `settings.n.convertPath`, then `settings.node.convertPath`, or no conversion.
The replacement uses JavaScript regular expressions and capture substitutions.

## Differences from upstream

- **JavaScript files without module syntax.** In `.js` files outside a
  TypeScript project, `sourceType: 'module'` can still miss `await (value)`,
  `await [value]`, `await ({ value })`, and `` await `value` ``.
  Add `export {}` or use an `.mjs` file to make these expressions checked.
- **Awaiting a negation.** In JavaScript, `await !value;` can produce a syntax
  error instead of this rule's diagnostic. Write `await (!value);` in a file
  with `export {}` or an `.mjs` extension to check it.
- **Parenthesized operands in computed names.** A top-level expression such as
  `const object = { [await (keyPromise)]() {} };` can be missed. Use
  `[await keyPromise]`, or move `const key = await (keyPromise);` before the
  object or class declaration, to make the await visible to this rule.
- **Overlapping path conversions.** Object-form `convertPath` patterns use
  alphabetical order. If both `**` and `src/**` match, `**` wins even when you
  wrote `src/**` first; upstream uses insertion order. This can change whether
  the resulting file is published or treated as an executable. Use the array
  form when priority matters.
- **Publication paths.** Package-root files such as `README.js` remain checked
  when linting from another directory. A filename such as `..hidden.js` remains
  inside its package. `convertPath` keeps the source package's publication
  boundary: with `"files": ["lib"]`, a target such as `lib/nested/private.js`
  stays published even if `lib/nested/package.json` lists only `public.js`.
  Subdirectory `.npmignore` files, or `.gitignore` when absent, still apply;
  they can exclude that target. Upstream can use the nested package's metadata
  with a path relative to the outer package and incorrectly skip the check.
  For a package directory named `Pkg`, a converted path `../pkg/lib/a.js`
  remains inside the package on a case-insensitive filesystem.
- **Unusual filename patterns.** In `files`, `.npmignore`, and `.gitignore`,
  `[!b]oo.js` matches `foo.js` but not `boo.js`, and `cli\?` matches a literal
  question mark. Upstream can select different files for these patterns. Each
  `files` entry also stays one pattern: `"lib/foo.js\nbar.js"` does not publish
  `lib/foo.js`, whereas upstream treats the newline as a separator. Use separate
  array entries for separate filenames. Further escaping and whitespace cases
  are described in [hashbang's file-selection notes](/rules/node/hashbang).
- **Malformed patterns and metadata.** An unclosed bracket such as `[cli.js`
  is matched literally. Upstream matches nothing for this pattern in an ignore
  file and can throw when it appears in `files`. A brace range with a zero
  step, such as `lib/cli{1..3..0}.js`, includes `lib/cli1.js` here but not upstream;
  use a positive step or list filenames explicitly. An invalid `convertPath`
  regular expression such as `[` skips the rule for that file, and non-string
  `bin` entries are ignored. Upstream throws in those last two cases. Correct
  the expression or executable path so the intended files are checked.

## Original documentation

- [eslint-plugin-n: no-top-level-await](https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/docs/rules/no-top-level-await.md)
- [Source code](https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/lib/rules/no-top-level-await.js)
