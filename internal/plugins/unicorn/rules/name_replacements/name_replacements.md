<!-- cspell:ignore pascalcase -->

# name-replacements

## Rule Details

Enforces descriptive replacements for variable, property, import, and filename names. The built-in replacements primarily expand common abbreviations, such as `err` to `error`, `btn` to `button`, and `props` to `properties`.

Examples of **incorrect** code for this rule:

```javascript
const err = new Error();
class Btn {}
```

Examples of **correct** code for this rule:

```javascript
const error = new Error();
class Button {}
```

Properties are not checked by default:

```javascript
const levels = { err: 0 };
this.evt = 'click';
```

Variables with exactly one available replacement are automatically renamed, including their references. Ambiguous variables and checked properties provide editor suggestions. Exported declaration names, exported property aliases, TypeScript parameter properties, and parameters with an attached JSDoc `@param` comment are reported without an edit when a rename could leave another name stale.

## Options

### `replacements`

Type: `Record<string, false | Record<string, boolean>>`

Extends the default replacements. Set a replacement to `false` to disable it, or set a complete discouraged name to `false` to disable all of that name's default replacements.

```json
{
  "unicorn/name-replacements": [
    "error",
    {
      "replacements": {
        "e": { "event": false },
        "res": false,
        "usr": { "user": true }
      }
    }
  ]
}
```

Lowercase entries match complete camelcase and pascalcase words. A camelcase entry such as `errCb` matches only that complete identifier (and `ErrCb`), not a word inside a longer identifier.

### `extendDefaultReplacements`

Type: `boolean`

Default: `true`

Set this to `false` to replace the defaults instead of extending them.

### `allowList`

Type: `Record<string, boolean>`

Extends the built-in case-sensitive allow list. It matches complete names only.

```json
{
  "unicorn/name-replacements": [
    "error",
    { "allowList": { "getInitialProps": true } }
  ]
}
```

### `extendDefaultAllowList`

Type: `boolean`

Default: `true`

Set this to `false` to replace the default allow list.

### `checkDefaultAndNamespaceImports`

Type: `"internal" | boolean`

Default: `"internal"`

- `"internal"` checks default imports, namespace imports, and static `require()` bindings only when the module specifier begins with `.` or `/`.
- `true` checks them for every module.
- `false` does not check them.

### `checkShorthandImports`

Type: `"internal" | boolean`

Default: `"internal"`

Controls unaliased named imports in the same way as `checkDefaultAndNamespaceImports`.

### `checkShorthandProperties`

Type: `boolean`

Default: `false`

Checks shorthand bindings in object destructuring, such as `const {err} = value`. A shorthand binding with a default value, such as `const {err = null} = value`, is checked regardless because its ESTree value is an assignment pattern rather than a shorthand identifier.

### `checkProperties`

Type: `boolean`

Default: `false`

Checks static property definitions and property writes. Property reads such as `object.err` remain allowed.

### `checkVariables`

Type: `boolean`

Default: `true`

Controls variable, parameter, type, class, and import binding checks.

### `checkFilenames`

Type: `boolean`

Default: `true`

Checks the basename of each linted file while preserving its final extension.

### `ignore`

Type: `string[]`

Default: `[]`

Adds JavaScript regular-expression patterns that suppress a complete name. The built-in patterns cover common conventions such as `i18n`, `l10n`, `a11y`, `e2e`, `jQuery`, and `vite-env`.

## Differences from ESLint

- Native rslint currently runs this rule on JavaScript and TypeScript source files. The upstream rule also checks JSON, JSONC, JSON5, YAML, TOML, CSS, HTML, Vue, and Markdown identifiers and filenames; those non-JavaScript language hooks are not available. Apply the rule only to JavaScript and TypeScript files when sharing a configuration with ESLint.
- JavaScript `RegExp` objects do not cross rslint's native configuration boundary. Write `ignore` entries as pattern strings. Flags on a regular-expression literal cannot be represented directly; express equivalent matching in the pattern where possible.
- A hidden or non-code filename that the TypeScript program does not load cannot produce a native filename diagnostic. JavaScript and TypeScript files selected for linting are checked normally.

## Original Documentation

- [eslint-plugin-unicorn: name-replacements](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v77.0.0/docs/rules/name-replacements.md)
- [Source code](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v77.0.0/rules/name-replacements.js)
