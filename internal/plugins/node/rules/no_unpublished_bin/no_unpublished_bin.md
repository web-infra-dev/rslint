# no-unpublished-bin

Disallow package executables excluded by publication settings.

This rule can be disabled when publishing with npm 10 or later, as noted in
the upstream documentation.

## Rule details

For each linted file, this rule finds the nearest `package.json` and checks
whether the file matches its `bin` field. It reports the complete file if the
executable is excluded by `files`, `.npmignore`, or the applicable `.gitignore`.
It does not check whether unvisited executable files exist and provides no fixes.

This package configuration is **incorrect** when linting `bin/cli.js`:

```json
{
  "bin": "bin/cli.js",
  "files": ["lib"]
}
```

Include the executable to make the configuration **correct**:

```json
{
  "bin": "bin/cli.js",
  "files": ["lib", "bin"]
}
```

Both string and object `bin` entries are supported. Entries without `.js` or
ending at a directory also match the corresponding `.js` or `/index.js` file.
An existing `.npmignore` replaces `.gitignore`; a `files` array also disables
`.gitignore` fallback. A `.npmignore` exclusion still applies to entries in
`files`. The package's `main` file and always-published metadata are exempt.

## Options

```javascript
export default [
  {
    plugins: ['node'],
    rules: {
      'node/no-unpublished-bin': ['error', {
        convertPath: [{
          include: ['src/**'],
          exclude: ['**/*.spec.ts'],
          replace: ['^src/(.*)\\.ts$', 'dist/$1.js'],
        }],
      }],
    },
  },
];
```

`convertPath` maps paths relative to the package directory before checking
`bin` and publication settings. The first matching array entry wins. Each
entry has `include`, optional `exclude`, and a JavaScript regular expression
and replacement pair in `replace`. For example, the configuration above checks
`src/cli.ts` as `dist/cli.js`.

The object form is also supported:
`{ "src/**": ["^src/(.*)\\.ts$", "dist/$1.js"] }`.
When the option is omitted, the rule uses `settings.n.convertPath`, then
`settings.node.convertPath`, or leaves the path unchanged. Omit `convertPath`
instead of setting it to `null`, which the upstream schema also rejects.

## Differences from upstream

The following configurations can produce different results from
`eslint-plugin-n` v18.3.0:

- Overlapping object-form `convertPath` patterns use alphabetical order instead
  of property order. For example, `**` wins over `src/**` even if `src/**` was
  written first. Use the array form when priority matters.
- Conversion targets starting with `/` resolve from the filesystem root. For
  example, `/dist/cli.js` does not match a package-local `dist/cli.js` bin entry.
  Upstream joins that target to the package directory and can throw when
  checking publication patterns. Use package-relative replacements.
- Published package-root metadata such as `README.js` is exempt regardless of
  the working directory. A published `..hidden.js` is treated as inside the
  package. When conversion enters a nested package, publication patterns are
  relative to that package. Upstream can report these files incorrectly.
- Publication patterns can select different files when they contain negated
  character classes, escaped characters, whitespace, invalid brackets, or
  zero-step brace ranges. For example,
  `"files": ["bin/[!b]oo.js"]` includes `bin/foo.js` in rslint; upstream reports
  that executable as unpublished. Each `files` entry stays one pattern:
  `"files": ["lib/foo.js\nbar.js"]` excludes `lib/foo.js` in rslint but includes
  it upstream. List separate filenames in separate entries. An unclosed
  `[cli.js` pattern matches that literal filename in rslint, while upstream
  matches nothing. Use closed brackets and positive brace steps. For escaped
  characters and unusual whitespace, see the additional filename examples in
  [hashbang](../hashbang/hashbang.md).
- An invalid conversion expression such as `[` skips this rule for the file;
  a non-string `bin` entry such as `{ "cli": false }` is ignored. Upstream
  throws in these cases. Correct the expression or use string executable paths.

## Original documentation

- [eslint-plugin-n: no-unpublished-bin](https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/docs/rules/no-unpublished-bin.md)
- [Source code](https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/lib/rules/no-unpublished-bin.js)
