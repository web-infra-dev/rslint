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
At the package root, `.npmignore` replaces `.gitignore`; a `files` array also
disables `.gitignore` fallback. Subdirectory ignore files still apply.
A `.npmignore` exclusion still applies to entries in
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
Use the array form to specify priority when patterns overlap.
When the option is omitted, the rule uses `settings.n.convertPath`, then
`settings.node.convertPath`, or leaves the path unchanged. Omit `convertPath`
instead of setting it to `null`, which the upstream schema also rejects.

## Differences from upstream

Compared with `eslint-plugin-n` v18.3.0, these inputs can change whether an
executable is reported as unpublished:

- **Converted executable paths.** A replacement such as `/dist/cli.js` names
  an absolute path, so it does not identify the package's `dist/cli.js`
  executable. Upstream treats it as a package-local path and can stop with an
  error. Use `dist/cli.js` to refer to the executable inside the package.
- **Published package files.** Package-root metadata such as `README.js` is
  not reported, even when linting from another directory. A filename such as
  `..hidden.js` is still inside the package and is not reported when included
  in `files`. Upstream can incorrectly report these files as unpublished.
- **Executables inside nested packages.** Suppose a package declares
  `"bin": "nested/cli.js"` and `"files": ["nested"]`, and `convertPath` maps
  `src/cli.js` to `nested/cli.js`. rslint does not report that executable,
  even if `nested/package.json` has `"files": []`; upstream reports it as
  unpublished. The package declaring `bin` determines which files it publishes.
- **Subdirectory ignore files.** With `"files": ["lib"]`, a `lib/.npmignore`
  containing `cli.js` excludes the executable `lib/cli.js` in rslint. Upstream
  checks only the package-root ignore file and does not report that executable.
- **Filename case.** On a case-insensitive filesystem, `"bin": "bin/CLI.js"`
  also selects `bin/cli.js`. With `"files": []`, rslint reports that file;
  upstream misses it. If `main` names the same file with different case,
  rslint keeps its publication exemption. On a case-sensitive filesystem,
  these filenames remain distinct.
  The same applies to aliases: `"bin": "bin/cli"` selects `BIN/CLI.JS` and
  `BIN/CLI/INDEX.JS` on a case-insensitive filesystem.
- **Excluding and escaping filename characters.** With
  `"files": ["bin/[!b]oo.js"]`, rslint includes `bin/foo.js` and does not report
  it; upstream reports it as unpublished. Escaped wildcards name literal
  characters: `"files": ["bin/cli\\*"]` includes the filename `bin/cli*`, but
  rslint still reports `bin/cli.js`; upstream includes both. The same matching
  differences apply to `.npmignore` and `.gitignore`. More filename examples
  appear in [hashbang](../hashbang/hashbang.md).
- **Whitespace in filenames.** With `"files": ["lib/foo.js\nbar.js"]`, rslint
  reports the executable `lib/foo.js`; upstream includes it. The newline is
  part of one filename pattern in rslint. List separate filenames in separate
  entries. Tabs and non-breaking spaces can also change which files match;
  see the [hashbang examples](../hashbang/hashbang.md).
- **Malformed filename patterns.** With `"files": ["[cli.js"]`, rslint includes
  the literal filename `[cli.js`; upstream throws an error for the malformed
  pattern. With `"files": ["lib/cli{1..3..0}.js"]`, rslint includes `lib/cli1.js`; upstream
  reports it. Use closed brackets and positive brace steps.
- **Invalid conversion expressions or executable entries.** An expression
  such as `[` causes rslint to skip this rule for the file. A `bin` entry such
  as `{ "cli": false }` is ignored. Upstream stops with an error in these
  cases. Correct the expression or use string executable paths so the file
  can be checked.

## Original documentation

- [eslint-plugin-n: no-unpublished-bin](https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/docs/rules/no-unpublished-bin.md)
- [Source code](https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/lib/rules/no-unpublished-bin.js)
