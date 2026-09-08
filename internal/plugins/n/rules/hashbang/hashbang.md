# hashbang

Require the correct hashbang for package executables.

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

Enable the bundled `n` plugin and configure any of these options:

```javascript
export default [
  {
    plugins: ['n'],
    rules: {
      'n/hashbang': ['error', {
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

- **Overlapping path conversions.** If several object-form `convertPath`
  patterns match a source file, rslint selects the first pattern in alphabetical
  order; upstream uses the order of the JavaScript object properties. For
  example, `**` wins over `src/**` in rslint even when you wrote `src/**` first.
  This can change whether the converted path matches a `bin` entry and needs a
  hashbang. Use the array form of `convertPath` to specify the priority.
- **Emoji in Git ignore patterns.** With `additionalExecutables: ["??.js"]`,
  rslint does not require a hashbang in `😀.js`, while upstream does. Similar
  differences affect plain `package.json` `files` patterns and ignore files.
  List the file name explicitly, or use `*` for names of any length. Extended
  `files` patterns such as `lib/@(??).js` match emoji names as upstream does.
- **Similar-looking characters in executable names.** The Latin letter `K`
  (`U+004B`) and the Kelvin sign `K` (`U+212A`) are different characters, even
  when they look identical in your font. With `additionalExecutables: ["K.js"]`,
  rslint also requires a hashbang in a file named `"\u212A.js"`; upstream does
  not select that file as an additional executable. Similar matching differences
  can affect which files are skipped by `ignoreUnpublished`. Use ordinary Latin
  letters in these file names to avoid the ambiguity.
- **Zero-step brace ranges.** With `"files": ["lib/cli{1..3..0}.js"]` in
  `package.json`, rslint treats `lib/cli1.js` as published; upstream does not.
  This can change whether `ignoreUnpublished` skips its hashbang check. Use a
  positive step, such as `lib/cli{1..3..1}.js`, or list the file names explicitly.
- **Fixing an empty hashbang or one without LF.** For an executable containing
  only `#!/usr/bin/env node` without a final newline, both tools report a
  missing hashbang. rslint replaces that header and adds LF; upstream inserts
  another hashbang before it, producing invalid JavaScript. rslint also replaces
  an empty `#!` header and headers ending in CR, U+2028 or U+2029 this way.
- **Invalid conversion expressions.** If a `convertPath` regular expression
  fails to compile, for example `[`, rslint skips this rule for the file and
  produces no hashbang diagnostic or fix. Upstream fails while loading the rule
  with a `SyntaxError`. Correct the expression to ensure the file is checked.
- **Non-string executable entries.** If `package.json` contains
  `"bin": { "cli": false }`, rslint ignores that entry; upstream throws a
  `TypeError`. Use a string path for each executable entry.

## Original Documentation

- [eslint-plugin-n: hashbang](https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/docs/rules/hashbang.md)
- [Source code](https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/lib/rules/hashbang.js)
