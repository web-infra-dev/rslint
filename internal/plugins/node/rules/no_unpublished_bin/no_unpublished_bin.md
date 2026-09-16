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
A `.npmignore` exclusion still applies to entries in `files`. The package's `main` file and always-published metadata are exempt.

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
`src/cli.ts` as `dist/cli.js`. Invalid regular expressions in rule options
are rejected as configuration errors.

The object form is also supported:
`{ "src/**": ["^src/(.*)\\.ts$", "dist/$1.js"] }`.
Use the array form to specify priority when patterns overlap.
When the option is omitted, the rule uses `settings.n.convertPath`, then
`settings.node.convertPath`, or leaves the path unchanged. Omit `convertPath`
instead of setting it to `null`, which the upstream schema also rejects.

## Differences from upstream

Compared with `eslint-plugin-n` v18.3.0:

- **Publication checks:** rslint uses the declaring package's settings and
  respects subdirectory ignore files. It avoids upstream false reports for
  root metadata such as `README.js` and included files such as `..hidden.js`.
- **Filename case:** `bin` (including `.js` and `/index.js` aliases) and `main`
  follow the filesystem's case rules. Upstream can miss case-only matches.
- **Converted paths:** absolute replacements remain absolute; upstream treats
  them as package-relative and may fail. Use relative paths for package files.
- **Filename patterns:** `[!b]` excludes `b`, and `\*` matches a literal `*`.
  Newlines, tabs and non-breaking spaces in a `files` entry stay in that
  pattern. Upstream can include or exclude different files.
- **Invalid input:** unclosed `[` matches literally, and a zero brace step
  behaves as one. Non-string `bin` entries are ignored; invalid conversion
  regexes in shared settings skip the check. Upstream may fail or produce
  different matches for these inputs.

## Original documentation

- [eslint-plugin-n: no-unpublished-bin](https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/docs/rules/no-unpublished-bin.md)
- [Source code](https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/lib/rules/no-unpublished-bin.js)
