# no-sync

Disallow synchronous calls identified by names ending in `Sync`.

Synchronous I/O can block other work in a Node.js application. Prefer an
asynchronous API where the surrounding code can wait for its result.

## Examples

```js
// Incorrect
fs.readFileSync(path);
readFileSync(path);
readFileSync.call(null, path);

// Correct
await fs.promises.readFile(path);
```

The rule checks names without requiring a Node.js import. It also checks direct
identifier arguments, such as `consume(readFileSync)`, and identifiers inside a
member-expression callee. String keys such as `fs['readFileSync']()` are allowed.
The rule provides no automatic fixes or suggestions.

## Options

### allowAtRootLevel

Set `allowAtRootLevel` to `true` to allow synchronous calls outside functions.
The default is `false`. Top-level blocks and class field initializers are also
outside functions; method bodies and function parameter defaults are inside.

```js
// With { allowAtRootLevel: true }
const config = fs.readFileSync(path); // Allowed

function reload() {
  return fs.readFileSync(path); // Reported
}
```

### ignores

The default is `[]`. Strings ignore exact identifier names:

```js
export default [{
  plugins: ['node'],
  rules: {
    'node/no-sync': ['error', { ignores: ['readFileSync'] }],
  },
}];
```

Object entries use TypeScript declaration information:

| Entry | Matches |
| --- | --- |
| `{ from: 'file', path: './helpers.ts' }` | Local declarations matching an optional file glob |
| `{ from: 'package', package: 'effect' }` | Declarations from an optional package name pattern |
| `{ from: 'lib' }` | TypeScript standard library declarations and intrinsic types |

Use `/` separators in file patterns on every platform. Backslashes escape glob
characters. Package patterns are JavaScript regular expressions in Unicode mode;
invalid patterns, such as `(`, are configuration errors.

Each object can include a `name` array. These names refer to the declared type,
including its containing class or interface, such as
`{ from: 'lib', name: ['CSSStyleSheet.replaceSync'] }`. Imported aliases use the
original declaration name. Omit `name` to ignore all matching declarations.

## Differences from upstream

- When type information is unavailable, object entries in `ignores` suppress
  nothing; string entries still apply. Upstream stops with an error in this case.
- An empty file pattern ignores no declarations. Upstream stops with an error.
- File patterns such as `./helpers.ts` also work on Windows and in working
  directories containing uppercase letters on case-insensitive file systems.
  Upstream can fail to ignore matching declarations in these cases.
- Directory checks distinguish siblings with a shared prefix. For example, with
  `typeRoots: ['./types']`, `{ from: 'file' }` can ignore declarations in
  `types-extra/helper.ts`; upstream incorrectly excludes them. A file glob cannot
  match declarations in a sibling `project-extra` directory when the working
  directory is `project`, even though upstream may accept them.
- Package patterns are validated before linting. Upstream may only report an
  invalid pattern when it encounters a matching call. Long Unicode property
  names such as `\p{Letter}` are unsupported; use `\p{L}` instead.

Some file globs match differently. The table shows whether each pattern ignores
an example declaration in a case-sensitive project directory. Prefer explicit
paths, `**/foo.ts`, or alternatives such as `**/*.{ts,tsx}` when sharing a
configuration with upstream.

| File pattern | Example declaration | rslint ignores | Upstream ignores |
| --- | --- | --- | --- |
| `./src/*.ts` | `src/foo.ts` | Yes | No |
| `**/[!a-z]*.ts` | `foo.ts` | No | Yes |
| `**/[^a-z]*.ts` | `.foo.ts` | No | Yes |
| `**/[[:digit:]]*.ts` | `1.ts` | No | Yes |
| `**/!(foo\|bar).ts` | `foobar.ts` | Yes | No |
| `!(**/foo.ts)` | `foo.ts` | Yes | No |
| `**/.*/foo.ts` | `foo.ts` | Yes | No |
| `**/foo.(ts)` | `foo.ts` | No | Yes |
| `**/[foo].ts` | `[foo].ts` | No | Yes |
| `**/file{01..03}.ts` | `file01.ts` | Yes | No |
| `**/file{1..3..2}.ts` | `file2.ts` | No | Yes |
| `**//foo.ts` | `foo.ts` | Yes | No |
| `**/foo.ts/**` | `foo.ts` | No | Yes |

## Original documentation

- [eslint-plugin-n: no-sync](https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/docs/rules/no-sync.md)
- [Source code](https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/lib/rules/no-sync.js)
