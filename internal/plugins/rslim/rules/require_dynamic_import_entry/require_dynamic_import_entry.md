# rslim/require-dynamic-import-entry

Report every runtime `import(...)`, bare `require(...)` call, and TypeScript `import x = require(...)` declaration so each dependency is reviewed before Rslim optimizes the project. This rule only detects the syntax. It does not resolve modules, read Rslim configuration, or require type information. Static ES imports, `require.resolve(...)`, and TypeScript `import(...)` types are not reported. Type-only import-equals declarations are ignored.

For each diagnostic:

1. Determine which modules Rspack can load from the import or require expression. Include every possible target of a variable or template path and account for module top-level side effects.
2. Check whether Rslim will preserve the runtime code of each target. Add at-risk source files to Rslim `entries`; marking one declaration with `@entry` may not preserve the module's top-level code.
3. After verifying every target, suppress this specific dependency and record the reason. If the targets cannot be determined, leave the diagnostic open for investigation.

```ts
// rslint-disable-next-line rslim/require-dynamic-import-entry -- All possible route modules are in Rslim entries.
const route = await import(`./routes/${name}.js`);
```

The same review applies to `require(...)` and TypeScript import-equals declarations:

```ts
// rslint-disable-next-line rslim/require-dynamic-import-entry -- Verified the loaded module is preserved.
const plugin = require('./plugin');
```

The diagnostic covers the entire `import(...)` call. For a multiline call, place the suppression immediately before the line containing `import`:

```ts
// rslint-disable-next-line rslim/require-dynamic-import-entry -- Verified all possible route modules.
const route = await import(
  path
);
```

The diagnostic provides the review steps instead of an automatic code suggestion because adding an ignore comment before checking the candidate modules could hide an unsafe import.

The diagnostic covers the entire `import(...)` or `require(...)` call, or the entire TypeScript import-equals declaration. Only files checked by Rslint can produce it. Calls inside unchecked third-party dependencies require a separate review.
