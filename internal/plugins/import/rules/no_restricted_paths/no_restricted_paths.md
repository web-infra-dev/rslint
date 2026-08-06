# no-restricted-paths

## Rule Details

Some projects contain files that are not meant to run in the same environment. A web application, for example, may hold server-only code next to browser code, and importing the server code from the client bundle would ship it to the browser.

This rule lets you declare restricted zones: a `target` set of files, and a `from` set of files those targets are not allowed to import.

The rule takes one option object with a list of `zones` and an optional `basePath` used to resolve the relative paths inside each zone. `basePath` defaults to the current working directory.

Each zone accepts:

- `target` — which files belong to the zone. A directory path (matching everything inside it recursively), a glob pattern, or an array of either.
- `from` — which files the zone may not import. A directory path, a glob pattern, or an array of only directory paths or only glob patterns.
- `except` — optional. Imports that are allowed even though `from` covers them. When `from` is a directory path, each entry is resolved relative to `from` and must stay inside it. When `from` is a glob pattern, each entry must be a glob pattern too.
- `message` — optional. Appended to the reported message.

`from` is matched against the resolved path of the imported file, not against the specifier text as written in the source.

Given this folder structure:

```
.
├── client
│   ├── foo.ts
│   └── baz.ts
└── server
    └── bar.ts
```

Examples of **incorrect** code for this rule:

```json
{ "import/no-restricted-paths": ["error", { "zones": [{ "target": "./client", "from": "./server" }] }] }
```

```javascript
// client/foo.ts
import bar from '../server/bar';
```

Examples of **correct** code for this rule:

```json
{ "import/no-restricted-paths": ["error", { "zones": [{ "target": "./client", "from": "./server" }] }] }
```

```javascript
// server/bar.ts
import baz from '../client/baz';
```

### `except`

Given this folder structure:

```
.
└── server
    ├── one
    │   ├── a.ts
    │   └── b.ts
    └── two
        └── a.ts
```

Examples of **incorrect** code for this rule:

```json
{
  "import/no-restricted-paths": [
    "error",
    { "zones": [{ "target": "./server/one", "from": "./server", "except": ["./one"] }] }
  ]
}
```

```javascript
// server/one/a.ts
import a from '../two/a';
```

Examples of **correct** code for this rule:

```json
{
  "import/no-restricted-paths": [
    "error",
    { "zones": [{ "target": "./server/one", "from": "./server", "except": ["./one"] }] }
  ]
}
```

```javascript
// server/one/a.ts
import b from './b';
```

### `basePath`

Relative `target` and `from` paths are resolved against `basePath`, and a relative `basePath` is itself resolved against the current working directory. `except` entries stay relative to their zone's `from`.

```json
{
  "import/no-restricted-paths": [
    "error",
    { "basePath": "./src", "zones": [{ "target": "./client", "from": "./server" }] }
  ]
}
```

### `message`

```json
{
  "import/no-restricted-paths": [
    "error",
    {
      "zones": [
        { "target": "./client", "from": "./server", "message": "Use the API client instead." }
      ]
    }
  ]
}
```

```javascript
// client/foo.ts
import bar from '../server/bar';
```

reports `Unexpected path "../server/bar" imported in restricted zone. Use the API client instead.`

## Differences from ESLint

- Glob patterns support `*`, `**`, `?`, `[abc]` character classes and `{a,b}` alternatives. Extended glob syntax — `!(a)`, `@(a|b)`, `+(a)`, `?(a)`, `*(a)` — is matched as the literal text it is written as, so `./src/?(server)/**/*` covers a directory named `?(server)` rather than one named `server`. Write `{server,shared}` instead of `@(server|shared)`, and name the directories you want to cover instead of excluding one with `!(...)`.
- A `*` matches path segments that begin with a dot, so `./src/*` covers `./src/.hidden.ts` as well.
- A zone is matched against the file a specifier resolves to, and a specifier resolves through the module resolution TypeScript recorded for it. TypeScript records `require('./x')` calls in JavaScript files, so in a TypeScript file write the import as an `import` declaration or a dynamic `import('./x')` for the zone to cover it.

## Original Documentation

- [import/no-restricted-paths](https://github.com/import-js/eslint-plugin-import/blob/main/docs/rules/no-restricted-paths.md)
