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

Each object can include a `name` array. These names refer to the declared type,
including its containing class or interface, such as
`{ from: 'lib', name: ['CSSStyleSheet.replaceSync'] }`. Imported aliases use the
original declaration name. Omit `name` to ignore all matching declarations.

## Differences from upstream

- When type information is unavailable, object entries in `ignores` suppress
  nothing; string entries still apply. Upstream stops with an error in this case.
- The file pattern `**/foo.(ts)` matches parentheses literally, so it does not
  ignore declarations in `foo.ts`. Upstream also accepts this pattern for
  `foo.ts`. Use `**/foo.ts` or `**/foo.{ts,tsx}` for the same result in both.
- An empty file pattern ignores no declarations. Upstream stops with an error.

## Original documentation

- [eslint-plugin-n: no-sync](https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/docs/rules/no-sync.md)
- [Source code](https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/lib/rules/no-sync.js)
