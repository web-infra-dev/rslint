# rslim/require-dynamic-import-entry

Report runtime `import(...)` calls whose first argument is not a string literal or a template literal without substitutions, so dependencies that TypeScript cannot analyze are reviewed before Rslim optimizes the project. Variable paths, concatenation, and template literals with substitutions are reported, even when a local variable is declared with `const`.

Bare `require(...)` calls are also reported, except in JavaScript files when the call has exactly one string literal or template literal without substitutions as its argument. TypeScript collects these JavaScript dependencies, but does not collect ordinary `require(...)` calls in TypeScript files, even with a literal argument. TypeScript `import x = require(...)` declarations are ignored because TypeScript analyzes their module references.

This rule only detects the syntax. It does not resolve modules, read Rslim configuration, or require type information. Static ES imports, `require.resolve(...)`, and TypeScript `import(...)` types are not reported.

These analyzable imports do not need a manual review through this rule:

```ts
const module = await import('./module');
const data = await import(`./data.json`, { with: { type: 'json' } });
import plugin = require('./plugin');
```

For each diagnostic:

1. Determine which modules Rspack can load from the import or require expression. Include every possible target of a variable or template path and account for module top-level side effects.
2. Check whether Rslim will preserve the runtime code of each target. Add at-risk source files to Rslim `entries`; marking one declaration with `@entry` may not preserve the module's top-level code.
3. After verifying every target, suppress this specific dependency and record the reason. If the targets cannot be determined, leave the diagnostic open for investigation.

```ts
// rslint-disable-next-line rslim/require-dynamic-import-entry -- All possible route modules are in Rslim entries.
const route = await import(`./routes/${name}.js`);
```

The same review applies to reported `require(...)` calls, including literal calls in TypeScript files:

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

The diagnostic covers the entire `import(...)` or `require(...)` call. Only files checked by Rslint can produce it. Calls inside unchecked third-party dependencies require a separate review.
