# unambiguous

## Rule Details

Require files configured with `languageOptions.sourceType: 'module'` to contain
a top-level import or export declaration. Files configured as `script` or
`commonjs` are ignored.

Examples of **incorrect** code for this rule:

```javascript
(function initialize() {
  console.log('ready');
})();
```

Examples of **correct** code for this rule:

```javascript
import './initialize.js';
```

```javascript
export function initialize() {
  console.log('ready');
}
```

```javascript
(function initialize() {
  console.log('ready');
})();

export {};
```

An empty `export {}` can mark a file that only has side effects as a module.
Type-only imports and exports, exported TypeScript declarations, and
TypeScript `export =` also satisfy this rule.

The check follows the upstream declaration-based heuristic. Dynamic `import()`,
`import.meta`, top-level `await`, CommonJS assignments, bare TypeScript
`import =` aliases, and `export as namespace` do not satisfy it. Exports inside
a namespace or an ambient module do not count as top-level exports. Empty
module files are also reported.

The rule does not provide automatic fixes or suggestions.

## Options

This rule has no options.

## When Not To Use It

Disable this rule if your environment determines the module format without
examining import or export declarations.

## Original Documentation

- [eslint-plugin-import: unambiguous](https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/docs/rules/unambiguous.md)
- [Source code](https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/src/rules/unambiguous.js)
