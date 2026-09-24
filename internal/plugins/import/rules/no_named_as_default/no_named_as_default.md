# no-named-as-default

## Rule Details

Reports a default import whose local name is also a named export of the imported module. This can indicate a missing pair of braces around a named import.

Given this module:

```javascript
// fruit.js
export default 'apple';
export const pear = 'pear';
```

Examples of **incorrect** code for this rule:

```javascript
import pear from './fruit.js';
```

Examples of **correct** code for this rule:

```javascript
import fruit from './fruit.js';
import { pear } from './fruit.js';
```

Unresolved and ignored modules, modules without a default export, and files using only CommonJS exports are skipped. TypeScript `export =` assignments and defaults supplied by `esModuleInterop` are also checked. Like upstream v2.32.0, the rule permits a name when both it and the default are explicit re-exports from the same file. Exporting the same local value under both names does not receive this exemption.

## Options

This rule has no options. It does not provide automatic fixes or suggestions.

## Differences from upstream

- When both `languageOptions.parserOptions.project` and `projectService` are
  disabled, only imported files also selected for linting can be checked. For
  example, linting `app.ts` alone does not check names exported by `lib.ts`.
  Enable either project option to check dependencies without linting them directly.
- Babel's experimental `export name from './module'` syntax is not supported.
  Use `export { default as name } from './module'` instead; this standard syntax
  is not checked by this rule, matching upstream.
- Syntax errors in imported files are not reported by this rule. For example,
  importing a file containing `return; export {};` produces a parse-error report
  upstream, but no report from this rule. Check the imported file's syntax separately.
- Quoted re-export names can produce reports that upstream misses. For example,
  given a module with `export default 1; export { foo as 'foo' } from './base'`,
  rslint reports `import foo from './module'`; upstream v2.32.0 does not.
  Quoting `default` or the name being re-exported also does not prevent rslint
  from checking for a name collision.
- rslint does not report an import just because its name matches a member of
  a re-exported namespace. For example, if `module.js`
  contains `export default 1; export * as names from './base'`, a named export
  `foo` in `base.js` does not make `import foo from './module'` an error.
  Upstream v2.32.0 reports this case.
- When `esModuleInterop` is omitted from `tsconfig.json`, configurations such as
  `module: 'nodenext'` can produce additional reports in rslint. For example, importing `foo`
  as the default from a module containing only `export const foo = 1` is reported.
  Upstream reports this case only with explicit `esModuleInterop: true`.
  With explicit `esModuleInterop: false`, neither tool reports this case.

## Original Documentation

- [eslint-plugin-import: no-named-as-default](https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/docs/rules/no-named-as-default.md)
- [Source code](https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/src/rules/no-named-as-default.js)
