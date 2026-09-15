# hashbang

Require the correct hashbang for package executables.

The `node` plugin ports rules from `eslint-plugin-n`.

## Rule Details

This rule finds the nearest `package.json` for each file. Files listed in its
`bin` field must start with `#!/usr/bin/env node`, without a Unicode BOM and with
an LF line ending. Other files must not have a hashbang. Files without a package
are left alone.

For a package with `"bin": "./bin/cli.js"`, this is **incorrect** in `bin/cli.js`:

```javascript
console.log('hello');
```

This is **correct**:

```javascript
#!/usr/bin/env node
console.log('hello');
```

In a regular library file, omit the hashbang. The rule provides automatic fixes
to insert, replace or remove the first line and to remove a BOM or CR character.
It recognizes hashbangs terminated by LF, including CRLF. It does not change
line endings in the rest of the file.

## Options

Enable the bundled `node` plugin and configure any of these options:

```javascript
export default [
  {
    plugins: ['node'],
    rules: {
      'node/hashbang': ['error', {
        ignoreUnpublished: false,
        additionalExecutables: ['scripts/cli.js'],
        executableMap: { '.ts': 'ts-node' },
        convertPath: [{
          include: ['src/**'],
          exclude: ['src/test/**'],
          replace: ['^src/(.*)\\.ts$', 'dist/$1.js'],
        }],
      }],
    },
  },
];
```

- `ignoreUnpublished` defaults to `false`. When enabled, use `package.json`
  `files`, `.npmignore` and the applicable `.gitignore` to skip unpublished files.
  Files matched by `additionalExecutables` are still checked.
- `additionalExecutables` defaults to `[]`. Its ordered Git ignore patterns
  select additional executable paths relative to the package directory and
  support `!` exclusions.
- `executableMap` maps file extensions to interpreter names. Extensions without
  an entry use `node`.
- `convertPath` maps source paths to published paths before checking `bin` and
  publication status. Each array entry has `include`, optional `exclude`, and
  `replace: [regularExpression, replacement]`. The first matching entry wins;
  replacement uses JavaScript capture substitutions such as `$1`. An object
  mapping patterns to replacement pairs is also supported. Omit the option to
  use `settings.n.convertPath`, then `settings.node.convertPath`, or no conversion.

## Differences from upstream

Compared with `eslint-plugin-n`, the following cases can change which files need
a hashbang, whether a file is checked, or the result of an automatic fix.

### Conversion priority

If several object-form `convertPath` patterns match, rslint uses alphabetical
pattern order; upstream uses object property order. For example, `**` wins over
`src/**` in rslint even if you wrote `src/**` first. The resulting path may match
a different `bin` entry or publication pattern. **Use the array form of
`convertPath` when priority matters.**

### Corrections to file selection and fixes

- **Checking published files.** With `ignoreUnpublished: true`, rslint keeps
  checking package-root files such as `README.js` when you lint from another
  directory. A name such as `..hidden.js` does not put a file outside its
  package. When `convertPath` points into a nested package, that package's
  `files` entries determine publication. Upstream can incorrectly skip the
  hashbang check in these cases.
- **Excluding or escaping filename characters.** With
  `additionalExecutables: ["[!b]oo.js"]` or `["[^b]oo.js"]`, rslint selects
  `foo.js` as an executable and excludes `boo.js`; upstream misses `foo.js`.
  With `["cli\\*"]`, rslint selects the literal filename `cli*`, while upstream
  also selects `cli.js`. Similarly, `["cli\\?"]` selects `cli?`, which upstream
  misses, and `["a\\b.js"]` selects `ab.js`, where upstream selects `a.js`.
  The same matching differences affect `files`, `.npmignore` and `.gitignore`.
- **Filenames containing whitespace.** Each `files` or `additionalExecutables`
  entry is one pattern, including any newline, tab or non-breaking space in it.
  For example, with `"files": ["lib/foo.js\nbar.js"]` and
  `ignoreUnpublished: true`, rslint skips `lib/foo.js`; upstream checks it.
  Put separate filenames in separate entries. Upstream can also ignore tabs
  and non-breaking spaces or match ordinary spaces instead. With
  `additionalExecutables: ["**/cli.js"]`, rslint selects `cli.js` inside a
  directory whose name contains a newline; upstream misses it. The directory
  name difference also affects `files`, `.npmignore` and `.gitignore`.
- **Fixing an incomplete first line.** If an executable contains only
  `#!/usr/bin/env node` without a final newline, both tools report a missing
  hashbang. rslint fixes it to a single header ending in LF; upstream duplicates
  the header and produces invalid JavaScript. rslint also replaces an empty
  `#!` or a header ending in CR, U+2028 or U+2029 with a valid header ending in LF.

### Invalid patterns and package metadata

- **Unclosed brackets or reversed ranges.** With
  `additionalExecutables: ["[cli.js"]`, rslint treats the literal filename
  `[cli.js` as executable; upstream selects nothing. A reversed range such as
  `[z-a]` also matches its literal text in rslint; upstream drops the invalid
  range and can still match the remaining characters. Close brackets and use
  ascending ranges such as `[a-z]` to select the intended files.
- **A brace range with step zero.** With
  `"files": ["lib/cli{1..3..0}.js"]` and `ignoreUnpublished: true`, rslint checks
  `lib/cli1.js`; upstream skips it. Use a positive step such as `{1..3..1}` or
  list the filenames separately.
- **An invalid `convertPath` regular expression.** An expression such as `[`
  causes rslint to skip this rule for the file without reporting or fixing a
  hashbang. Upstream stops with a `SyntaxError`. Correct the expression so the
  file is checked.
- **A non-string `bin` entry.** For `"bin": { "cli": false }`, rslint ignores
  that entry; upstream stops with a `TypeError`. Use a string path so the entry
  identifies an executable.

## Original Documentation

- [eslint-plugin-n: hashbang](https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/docs/rules/hashbang.md)
- [Source code](https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/lib/rules/hashbang.js)
